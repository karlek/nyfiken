// Package ini adds functionality to retrieve configuration from INI files.
package ini

import "net/url"
import "os"
import "strings"
import "time"

import "github.com/jteeuwen/ini"
import "github.com/karlek/nyfiken/page"
import "github.com/karlek/nyfiken/settings"
import "github.com/mewkiz/pkg/errorsutil"

const (
	// INI sections (i.e. [sectionName]).
	sectionSettings = "settings"
	sectionMail     = "mail"

	// INI field names.
	fieldInterval       = "interval"
	fieldBrowser        = "browser"
	fieldPortNum        = "portnum"
	fieldStrip          = "strip"
	fieldSleepStart     = "sleepstart"
	fieldRecvMail       = "recvmail"
	fieldFilePerms      = "fileperms"
	fieldNewline        = "newline"
	fieldSendMail       = "sendmail"
	fieldSendPass       = "sendpass"
	fieldSendAuthServer = "sendauthserver"
	fieldSendOutServer  = "sendoutserver"
	fieldSelection      = "sel"
	fieldRegexp         = "regexp"
	fieldNegexp         = "negexp"
	fieldThreshold      = "threshold"
	fieldHeader         = "header"
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
		fieldStrip:     true,
		fieldFilePerms: true,
	}

	// Error messages.
	errFieldNotExist          = "Field `%s` doesn't exist."
	errNoSectionSettings      = `No [` + sectionSettings + `] section found config.ini.`
	errNoSectionMail          = `No [` + sectionMail + `] section found in config.ini.`
	errInvalidMailAddress     = "Invalid mail: `%s`; correct syntax -> `name@domain.tld`."
	errInvalidHeader          = "Invalid header: `%s`; correct syntax -> `HeaderName: Value`."
	errMailAddressNotFound    = "Global receiving mail required."
	errMailAuthServerNotFound = "Sending mail authorization server required."
	errMailOutServerNotFound  = "Sending mail outgoing server required."
	errInvalidRandInterval    = "Invalid random interval: %s; correct syntax -> `duration duration`"
	errInvalidListDeclaration = "Use `<` instead of `=` for list values"
)

// Reads config file and updates settings.Global.
func ReadSettings() (err error) {

	// Parse config file.
	file := ini.New()
	err = file.Load(settings.ConfigPath)
	if err != nil {
		return errorsutil.ErrorfColor("%s", err)
	}

	config, settingExist := file.Sections[sectionSettings]
	mail, mailExist := file.Sections[sectionMail]
	if settingExist {
		err = parseSettings(config)
		if err != nil {
			return err
		}
	}
	if mailExist {
		err = parseMail(mail)
		if err != nil {
			return err
		}
	}

	return nil
}

// Parse ini settings section to global setting.
func parseSettings(config ini.Section) (err error) {
	for fieldName, _ := range config {
		if _, found := settingsFields[fieldName]; !found {
			return errorsutil.ErrorfColor(errFieldNotExist, fieldName)
		}
	}

	// Get time setting from INI.
	// If interval setting wasn't found, default value is 1 minute
	intervalStr := config.S(fieldInterval, settings.DefaultInterval.String())
	// Parse string to duration.
	settings.Global.Interval, err = time.ParseDuration(intervalStr)
	if err != nil {
		return errorsutil.ErrorfColor("%s", err)
	}

	// Set global file permissions.
	settings.Global.FilePerms = os.FileMode(config.I(fieldFilePerms, int(settings.DefaultFilePerms)))

	// Set global newline character.
	settings.Global.Newline = config.S(fieldNewline, settings.DefaultNewline)

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
			return errorsutil.ErrorfColor(errFieldNotExist, fieldName)
		}
	}

	// Set global receive mail.
	settings.Global.RecvMail = mail.S(fieldRecvMail, "")
	if settings.Global.RecvMail == "" {
		return errorsutil.ErrorfColor(errMailAddressNotFound)
	} else if !strings.Contains(settings.Global.RecvMail, "@") {
		return errorsutil.ErrorfColor(errInvalidMailAddress, settings.Global.RecvMail)
	}

	// Set global sender mail.
	settings.Global.SenderMail.Address = mail.S(fieldSendMail, "")
	if settings.Global.SenderMail.Address == "" {
		return errorsutil.ErrorfColor(errMailAddressNotFound)
	} else if !strings.Contains(settings.Global.SenderMail.Address, "@") {
		return errorsutil.ErrorfColor(errInvalidMailAddress, settings.Global.SenderMail.Address)
	}

	// Set global sender mail password.
	settings.Global.SenderMail.Password = mail.S(fieldSendPass, "")

	// Set global sender authorization server.
	settings.Global.SenderMail.AuthServer = mail.S(fieldSendAuthServer, "")
	if settings.Global.SenderMail.AuthServer == "" {
		return errorsutil.ErrorfColor(errMailAuthServerNotFound)
	}

	// Set global sender mail outgoing server.
	settings.Global.SenderMail.OutServer = mail.S(fieldSendOutServer, "")
	if settings.Global.SenderMail.OutServer == "" {
		return errorsutil.ErrorfColor(errMailOutServerNotFound)
	}

	return nil
}

// Reads pages file and returns a slice of pages.
func ReadPages() (pages []*page.Page, err error) {

	// Parse pages file.
	file := ini.New()
	err = file.Load(settings.PagesPath)
	if err != nil {
		return nil, errorsutil.ErrorfColor("%s", err)
	}

	// Loop through the INI sections ([section]) and parse page settings.
	for name, section := range file.Sections {
		// Skip global scope INI values since they are empty.
		if len(name) == 0 {
			continue
		}

		for fieldName, _ := range section {
			if _, found := siteFields[fieldName]; !found {
				return nil, errorsutil.ErrorfColor(errFieldNotExist, fieldName)
			}
		}

		// Sets the page variables.
		var p page.Page
		var pageSettings settings.Page

		// Make INI section ([http://example.org]) into url.URL.
		p.ReqUrl, err = url.Parse(name)
		if err != nil {
			return nil, errorsutil.ErrorfColor("%s", err)
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
			return nil, errorsutil.ErrorfColor("%s", err)
		}

		// Set individual mail address.
		pageSettings.RecvMail = section.S(fieldRecvMail, settings.Global.RecvMail)
		if pageSettings.RecvMail != "" && !strings.Contains(pageSettings.RecvMail, "@") {
			return nil, errorsutil.ErrorfColor(errInvalidMailAddress, pageSettings.RecvMail)
		}

		// Set individual header.
		headers := section.List(fieldHeader)
		m := make(map[string]string)
		for _, header := range headers {
			if strings.Contains(header, ":") {
				keyVal := strings.SplitN(header, ":", 2)
				m[strings.TrimSpace(keyVal[0])] = strings.TrimSpace(keyVal[1])
			} else {
				return nil, errorsutil.ErrorfColor(errInvalidHeader, header)
			}
		}
		pageSettings.Header = m

		// Set strip functions to use.
		pageSettings.StripFuncs = section.List(fieldStrip)
		if pageSettings.StripFuncs == nil {
			if _, found := section[fieldStrip]; found {
				return nil, errorsutil.ErrorfColor(errInvalidListDeclaration)
			}
		}

		p.Settings = pageSettings

		// Append to the pages array.
		pages = append(pages, &p)
	}

	return pages, nil
}

// Read config file which in turn updates settings.Global and returns a slice of
// all pages to scrape.
func ReadIni() (pages []*page.Page, err error) {

	// Read config.
	err = ReadSettings()
	if err != nil {
		return nil, err
	}

	// Read pages file.
	pages, err = ReadPages()
	if err != nil {
		return nil, err
	}

	return pages, nil
}
