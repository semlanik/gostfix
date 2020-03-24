/*
 * MIT License
 *
 * Copyright (c) 2020 Alexey Edelev <semlanik@gmail.com>
 *
 * This file is part of gostfix project https://git.semlanik.org/semlanik/gostfix
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and
 * to permit persons to whom the Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies
 * or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
 * PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
 * FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
 * OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package web

import (
	"bytes"
	template "html/template"
	ioutil "io/ioutil"
	"log"
)

const (
	IndexTemplateName      = "index.html"
	MailListTemplateName   = "maillist.html"
	DetailsTemplateName    = "details.html"
	ErrorTemplateName      = "error.html"
	LoginTemplateName      = "login.html"
	StatusLineTemplateName = "statusline.html"
	FoldersTemplateName    = "folders.html"
	MailNewTemplateName    = "mailnew.html"
	MailTemplateName       = "mailTemplate.eml"
	SignupTemplateName     = "signup.html"
	RegisterTemplateName   = "register.html"
	SettingsTemplateName   = "settings.html"
)

type Templater struct {
	indexTemplate      *template.Template
	mailListTemplate   *template.Template
	detailsTemplate    *template.Template
	errorTemplate      *template.Template
	loginTemplate      *template.Template
	signupTemplate     *template.Template
	registerTemplate   *template.Template
	statusLineTemplate *template.Template
	foldersTemaplate   *template.Template
	mailNewTemplate    *template.Template
	mailTemplate       *template.Template
	settingsTemplate   *template.Template
}

func NewTemplater(templatesPath string) (t *Templater) {
	t = nil
	index, err := parseTemplate(templatesPath + "/" + IndexTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	maillist, err := parseTemplate(templatesPath + "/" + MailListTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	details, err := parseTemplate(templatesPath + "/" + DetailsTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	errors, err := parseTemplate(templatesPath + "/" + ErrorTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	login, err := parseTemplate(templatesPath + "/" + LoginTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	statusLine, err := parseTemplate(templatesPath + "/" + StatusLineTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	folders, err := parseTemplate(templatesPath + "/" + FoldersTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	mailNew, err := parseTemplate(templatesPath + "/" + MailNewTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	mail, err := parseTemplate(templatesPath + "/" + MailTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	signup, err := parseTemplate(templatesPath + "/" + SignupTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	register, err := parseTemplate(templatesPath + "/" + RegisterTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	settings, err := parseTemplate(templatesPath + "/" + SettingsTemplateName)
	if err != nil {
		log.Fatal(err)
	}

	t = &Templater{
		indexTemplate:      index,
		mailListTemplate:   maillist,
		detailsTemplate:    details,
		errorTemplate:      errors,
		loginTemplate:      login,
		statusLineTemplate: statusLine,
		foldersTemaplate:   folders,
		mailNewTemplate:    mailNew,
		mailTemplate:       mail,
		signupTemplate:     signup,
		registerTemplate:   register,
		settingsTemplate:   settings,
	}
	return
}

func parseTemplate(path string) (*template.Template, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	return template.New("Index").Parse(string(content))
}

func (t *Templater) ExecuteIndex(data interface{}) string {
	return executeTemplateCommon(t.indexTemplate, data)
}

func (t *Templater) ExecuteMailList(data interface{}) string {
	return executeTemplateCommon(t.mailListTemplate, data)
}

func (t *Templater) ExecuteDetails(data interface{}) string {
	return executeTemplateCommon(t.detailsTemplate, data)
}

func (t *Templater) ExecuteError(data interface{}) string {
	return executeTemplateCommon(t.errorTemplate, data)
}

func (t *Templater) ExecuteLogin(data interface{}) string {
	return executeTemplateCommon(t.loginTemplate, data)
}

func (t *Templater) ExecuteSignup(data interface{}) string {
	return executeTemplateCommon(t.signupTemplate, data)
}

func (t *Templater) ExecuteRegister(data interface{}) string {
	return executeTemplateCommon(t.registerTemplate, data)
}

func (t *Templater) ExecuteStatusLine(data interface{}) string {
	return executeTemplateCommon(t.statusLineTemplate, data)
}

func (t *Templater) ExecuteFolders(data interface{}) string {
	return executeTemplateCommon(t.foldersTemaplate, data)
}

func (t *Templater) ExecuteNewMail(data interface{}) string {
	return executeTemplateCommon(t.mailNewTemplate, data)
}

func (t *Templater) ExecuteMail(data interface{}) string {
	return executeTemplateCommon(t.mailTemplate, data)
}

func (t *Templater) ExecuteSettings(data interface{}) string {
	return executeTemplateCommon(t.settingsTemplate, data)
}

func executeTemplateCommon(t *template.Template, values interface{}) string {
	buffer := &bytes.Buffer{}
	err := t.Execute(buffer, values)
	if err != nil {
		log.Printf("Could not execute template: %s", err)
	}
	return buffer.String()
}
