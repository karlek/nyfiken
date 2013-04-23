// Package ini adds functionality to retrieve configuration from INI files.
package ini

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jteeuwen/ini"
	"github.com/karlek/nyfiken/page"
	"github.com/karlek/nyfiken/settings"
	"github.com/mewkiz/pkg/errutil"
)

const (
	// INI sections (i.e. [sectionName]).
	sectionSettings = "settings"
	sectionMail     = "mail"

	// INI field names.
	fieldBrowser        = "browser"
	fieldFilePerms      = "fileperms"
	fieldHeader         = "header"
	fieldInterval       = "interval"
	fieldNegexp         = "negexp"
	fieldPortNum        = "portnum"
	fieldRecvMail       = "recvmail"
	fieldRegexp         = "regexp"
	fieldSelection      = "sel"
	fieldSendAuthServer = "sendauthserver"
	fieldSendMail       = "sendmail"
	fieldSendOutServer  = "sendoutserver"
	fieldSendPass       = "sendpass"
	fieldSleepStart     = "sleepstart"
	fieldStrip          = "strip"
	fieldThreshold      = "threshold"
)

var (
	// Valid fields in different sections
	siteFields = map[string]bool{
		fieldInterval:  true,
		fieldStrip:     true,
		fieldRecvMail:  true,
		fieldSelection: true,
		fieldRegexp:    true,
		fieldNegexp:    true,
		fieldThreshold: true,
		fieldHeader:    true,
	}
	mailFields = map[string]bool{
		fieldRecvMail:       true,
		fieldSendMail:       true,
		fieldSendPass:       true,
		fieldSendAuthServer: true,
		fieldSendOutServer:  true,
	}
	settingsFields = map[string]bool{
		fieldInterval:  true,
		fieldBrowser:   true,
		fieldPortNum:   true,
		fieldFilePerms: true,
	}

	// Error messages.
	ErrFieldNotExist          = "ini: field `%s` doesn't exist."
	ErrNoSectionSettings      = "ini: no [" + sectionSettings + "] section found config.ini."
	ErrNoSectionMail          = "ini: no [" + sectionMail + "] section found in config.ini."
	ErrInvalidMailAddress     = "ini: invalid mail: `%s`; correct syntax -> `name@domain.tld`."
	ErrInvalidHeader          = "ini: invalid header: `%s`; correct syntax -> `HeaderName: Value`."
	ErrInvalidRandInterval    = "ini: invalid random interval: %s; correct syntax -> `duration duration`."
	ErrMailAddressNotFound    = "ini: global receiving mail required."
	ErrMailAuthServerNotFound = "ini: sending mail authorization server required."
	ErrMailOutServerNotFound  = "ini: sending mail outgoing server required."
	ErrInvalidListDeclaration = "ini: use `<` instead of `=` for list values."
)

// Read config file which in turn updates settings.Global and returns a slice of
// all pages to scrape.
func ReadIni(configPath, pagesPath string) (pages []*page.Page, err error) {
	// Read config.
	err = ReadSettings(configPath)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Read pages file.
	pages, err = ReadPages(pagesPath)
	if err != nil {
		return nil, errutil.Err(err)
	}

	return pages, nil
}

// Reads config file and updates settings.Global.
func ReadSettings(configPath string) (err error) {
	// Parse config file.
	file := ini.New()
	err = file.Load(configPath)
	if err != nil {
		return errutil.Err(err)
	}

	config, settingExist := file.Sections[sectionSettings]
	mail, mailExist := file.Sections[sectionMail]
	if settingExist {
		err = parseSettings(config)
		if err != nil {
			return errutil.Err(err)
		}
	}
	if mailExist {
		err = parseMail(mail)
		if err != nil {
			return errutil.Err(err)
		}
	}

	return nil
}

// Parse ini settings section to global setting.
func parseSettings(config ini.Section) (err error) {
	for fieldName, _ := range config {
		if _, found := settingsFields[fieldName]; !found {
			return errutil.Newf(ErrFieldNotExist, fieldName)
		}
	}

	// Get time setting from INI.
	// If interval setting wasn't found, default value is 1 minute
	intervalStr := config.S(fieldInterval, settings.DefaultInterval.String())
	// Parse string to duration.
	settings.Global.Interval, err = time.ParseDuration(intervalStr)
	if err != nil {
		return errutil.Err(err)
	}

	// Set global file permissions.
	settings.Global.FilePerms = os.FileMode(config.I(fieldFilePerms, int(settings.DefaultFilePerms)))

	// Set port number.
	settings.Global.PortNum = config.S(fieldPortNum, settings.DefaultPortNum)

	// Set browser path.
	settings.Global.Browser = config.S(fieldBrowser, "")

	return nil
}

// Parse ini mail section to global setting.
func parseMail(mail ini.Section) (err error) {
	for fieldName, _ := range mail {
		if _, found := mailFields[fieldName]; !found {
			return errutil.Newf(ErrFieldNotExist, fieldName)
		}
	}

	// Set global sender mail.
	settings.Global.SenderMail.Address = mail.S(fieldSendMail, "")
	if settings.Global.SenderMail.Address == "" {
		return errutil.Newf(ErrMailAddressNotFound)
	} else if !strings.Contains(settings.Global.SenderMail.Address, "@") {
		return errutil.Newf(ErrInvalidMailAddress, settings.Global.SenderMail.Address)
	}

	// Set global sender mail password.
	settings.Global.SenderMail.Password = mail.S(fieldSendPass, "")

	// Set global sender authorization server.
	settings.Global.SenderMail.AuthServer = mail.S(fieldSendAuthServer, "")
	if settings.Global.SenderMail.AuthServer == "" {
		return errutil.Newf(ErrMailAuthServerNotFound)
	}

	// Set global sender mail outgoing server.
	settings.Global.SenderMail.OutServer = mail.S(fieldSendOutServer, "")
	if settings.Global.SenderMail.OutServer == "" {
		return errutil.Newf(ErrMailOutServerNotFound)
	}

	// Set global receive mail.
	settings.Global.RecvMail = mail.S(fieldRecvMail, "")
	if settings.Global.RecvMail == "" {
		return errutil.Newf(ErrMailAddressNotFound)
	} else if !strings.Contains(settings.Global.RecvMail, "@") {
		return errutil.Newf(ErrInvalidMailAddress, settings.Global.RecvMail)
	}

	return nil
}

// Reads pages file and returns a slice of pages.
func ReadPages(pagesPath string) (pages []*page.Page, err error) {

	// Parse pages file.
	file := ini.New()
	err = file.Load(pagesPath)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Loop through the INI sections ([section]) and parse page settings.
	for name, section := range file.Sections {
		// Skip global scope INI values since they are empty.
		if len(name) == 0 {
			continue
		}

		for fieldName, _ := range section {
			if _, found := siteFields[fieldName]; !found {
				return nil, errutil.Newf(ErrFieldNotExist, fieldName)
			}
		}

		var p page.Page
		var pageSettings settings.Page

		// Make INI section ([http://example.org]) into url.URL.
		p.ReqUrl, err = url.Parse(name)
		if err != nil {
			return nil, errutil.Err(err)
		}

		// Set CSS selector.
		pageSettings.Selection = section.S(fieldSelection, "")

		// Set regular expression string.
		pageSettings.Regexp = section.S(fieldRegexp, "")

		// Set "negexp" (negative regular expression) string which removes all
		// that matches it.
		pageSettings.Negexp = section.S(fieldNegexp, "")

		// Set threshold value.
		pageSettings.Threshold = section.F64(fieldThreshold, 0)

		// Set interval time.
		intervalStr := section.S(fieldInterval, settings.Global.Interval.String())
		// Parse string to duration.
		pageSettings.Interval, err = time.ParseDuration(intervalStr)
		if err != nil {
			return nil, errutil.Err(err)
		}

		// Set individual mail address.
		pageSettings.RecvMail = section.S(fieldRecvMail, settings.Global.RecvMail)
		if pageSettings.RecvMail != "" && !strings.Contains(pageSettings.RecvMail, "@") {
			return nil, errutil.Newf(ErrInvalidMailAddress, pageSettings.RecvMail)
		}

		// Set individual header.
		headers := section.List(fieldHeader)
		m := make(map[string]string)
		for _, header := range headers {
			if strings.Contains(header, ":") {
				keyVal := strings.SplitN(header, ":", 2)
				m[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
			} else {
				return nil, errutil.Newf(ErrInvalidHeader, header)
			}
		}
		pageSettings.Header = m

		// Set strip functions to use.
		pageSettings.StripFuncs = section.List(fieldStrip)
		if pageSettings.StripFuncs == nil {
			if _, found := section[fieldStrip]; found {
				return nil, errutil.Newf(ErrInvalidListDeclaration)
			}
		}

		p.Settings = pageSettings

		pages = append(pages, &p)
	}

	if pages == nil {
		return nil, errutil.Newf("no pages in %s", settings.PagesPath)
	}
	return pages, nil
}
