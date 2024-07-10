package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli"
)

type licenceType int

const (
	foss licenceType = iota
	open
	avail
	prop
)

var (
	year        int = time.Now().Year()
	fullname    string
	description string = "The standard legal tool for software."
	program     string = "befoss"
	email       string = "diogenesoft@protonmail.com"
	//go:embed licences/gplv3.txt
	gplv3 string
	//go:embed licences/mpl.txt
	mpl           string
	fset          Features // feature set of you program
	licence       string   // which licence you are using
	details       string
	defaultEditor string
)

type Program struct {
	Name string // name of your program
	Desc string // description of your program
}

type Licence struct {
	Content string
	Name    string
	Short   string
	TypeOf  licenceType
}

type Features struct {
	Licence      bool
	Contributing bool
	Prepend      bool
}

var licenceOptions = []huh.Option[string]{
	huh.NewOption("Gnu Public Licence, feel the hurd", "gpl"),
	huh.NewOption("MIT, KISS principled", "mit"),
	huh.NewOption("MPL, I HAVE MIDDLE GROUND!", "mpl"),
}

func main() {

	app := &cli.App{
		Name:  "befoss",
		Usage: "Befoss starts a form for you to choose your licence.",
		Action: func(*cli.Context) error {
			log.Info("It's dangerous to go alone, take a licence.")
			// let's just launch the huh command in interactive mode with this command.

			//  Pick an editor if the value does not exist in the environment.
			{
				var editorForm *huh.Form
				defaultEditor, ok := os.LookupEnv("EDITOR")
				if !ok {
					editorForm = huh.NewForm(
						huh.NewGroup(
							huh.NewSelect[string]().
								Title("Choose an editor, WARNING you will be judged. Pick wisely hero.").
								Options(
									huh.NewOption("EMACS, the editor for thy soul.", "emacs"),
									huh.NewOption("Helix deez", "hx"),
									huh.NewOption("VI VI VI", "vim"),
									huh.NewOption("uh, nano?", "nano"),
									huh.NewOption("Hai, Let's write shall we, kakoune it is.", "kak"),
								).
								Value(&defaultEditor),
						),
					)
				}

				err := editorForm.Run()
				if err != nil {
					log.Fatal(err)
				}
			}

			// Pick a licences and choose whether or not to create licence files, contributing guides, and prepending.
			{

				initForm := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("Choose your licence").
							Options(
								licenceOptions...,
							).
							Value(&licence),

						huh.NewConfirm().
							Title("Would you like to create a Licence file?").
							Value(&fset.Licence),
						huh.NewConfirm().
							Title("Would you like to create a contributing guide?").
							Value(&fset.Contributing),
						huh.NewConfirm().
							Title(fmt.Sprintf("Would you like to prepend %s to you files?", licence)).
							Value(&fset.Prepend),
					),
				)

				err := initForm.Run()
				if err != nil {
					log.Fatal(err)
				}
			}

			infoForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("What's your full name?").
						Value(&fullname).
						Validate(func(name string) error {
							if len(fullname) < 2 {
								return errors.New("Your name is too short")
							} else if len(fullname) > 30 {
								return errors.New("Your name is too long")
							}
							return nil
						}),
					huh.NewInput().
						Title("What's your program called anyway?").
						Value(&program),
					huh.NewInput().
						Title("I need the bloody description too!").
						Placeholder(description).
						Value(&description).
						Validate(func(name string) error {
							if len(fullname) < 15 {
								return errors.New("Your description is weak!")
							} else if len(fullname) > 60 {
								return errors.New("Your description's too long man, keep it simple geez.")
							}
							return nil
						}),
					huh.NewInput().
						Title("People will need to contact you in some way..., an email or mailing address would be nice").
						Placeholder(email).
						Value(&email),
					huh.NewText().
						Title("If you want to add somethin' extra, now's yer chance mate.").
						Value(&details),
				))

			err := infoForm.Run()
			if err != nil {
				log.Fatal(err)
			}

			licences := map[string]Licence{
				"mit": {
					Name: "Mit",
					Content: fmt.Sprintf(`
				
MIT License

Copyright (c) %d %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
			`, year, fullname),
					TypeOf: open,
				},
				"gpl": {
					Name: "gplv3", Content: gplv3,
					Short: fmt.Sprintf(`
    %s: %s
    Copyright (C) %d %s

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    You can contact me at %s.
				`, program, description, year, fullname, email),
					TypeOf: foss,
				},
				"mpl": {
					Name:    "mplv2",
					Content: mpl,
					TypeOf:  open,
				},
			}

			if fset.Licence {
				if licence == "custom" {
					var licenceTitle string
					var licenceShort string
					var licenceContent string

					customForm := huh.NewForm(
						huh.NewGroup(

							huh.NewInput().
								Title("What's your licence called?").
								Value(&licenceTitle).
								Validate(func(name string) error {
									if len(fullname) < 2 {
										return errors.New("Your name is too short")
									} else if len(fullname) > 40 {
										return errors.New("Your name is too long")
									}
									return nil
								}),
							huh.NewText().
								Editor(defaultEditor).
								Title("Gimme your lunch mon--I mean please enter a short licence description").
								Value(&licenceShort).
								Validate(func(name string) error {
									if len(fullname) < 2 {
										return errors.New("Your name is too short")
									} else if len(fullname) > 40 {
										return errors.New("Your name is too long")
									}
									return nil
								}),
							huh.NewText().
								Title("I want the loooongest description you got for your licence").
								Value(&licenceContent),
							huh.NewSelect[licenceType]().
								Title("What type of licences are we talkin' about here?").
								Options(
									huh.NewOption("Freedom! FOSS bossin', copying left and right.", foss),
									huh.NewOption("Open and permissive, let it be!", open),
									huh.NewOption("They can see but they can't touch. Source available", avail),
									huh.NewOption("After All, Why Not? Why Shouldn't I Keep It? It's mine.", prop),
								),
						))

					err := customForm.Run()
					if err != nil {
						log.Fatal(err)
					}
					licence = licenceTitle

					if entry, ok := licences[licence]; ok {
						entry.Content = licenceContent
						entry.Short = licenceShort
						entry.Name = licenceTitle
					} else {
						panic("Could not assign licence")
					}
				}
				// create the licence file
				file, err := os.Create("LICENCE")
				if err != nil {
					log.Fatal("Could not create licence file.", err)
				}
				defer file.Close()
				file.Write([]byte(licences[licence].Content))
				// Save the custom licence to the configuration so that next time the custom licence will be available as an option.
				// That would mean I would have to create Option and then add it to the licencesOptions
				licenceOptions = append(licenceOptions, huh.NewOption(licence, licence))

			} else {
				log.Info("What? You don't want a licence file?!")
			}

			if fset.Contributing {
				var contributorText string
				log.Info("Please note that the following text in your editor will be passed to a markdown file.")
				contributingForm := huh.NewForm(
					huh.NewGroup(

						huh.NewText().
							Title("Give us your standards for what a contribution looks like.").
							Editor(defaultEditor).
							Value(&contributorText),
					))

				err := contributingForm.Run()
				if err != nil {
					log.Fatal(err)
				}
				// create the contribute file
				file, err := os.Create("CONTRIBUTE.md")
				if err != nil {
					log.Fatal("Could not create contribute file.", err)
				}
				defer file.Close()
				file.Write([]byte(contributorText))
			}

			if fset.Prepend {
				var envrc string
				var extraText string
				prependForm := huh.NewForm(
					huh.NewGroup(

						huh.NewText().
							Title("Give us any text you would like to be prepended to your files in addition to licence information.").
							Editor(defaultEditor).
							Value(&extraText),
					))

				err := prependForm.Run()
				if err != nil {
					log.Fatal(err)
				}
				// TODO: prepend to the envrc file, CHANGE this to simply append,
				// I only need to prepend the info to other files not to envrc
				var file *os.File
				envrcBytes, err := os.ReadFile(".envrc")
				if err != nil {
					file, err = os.Create(".envrc")
					if err != nil {
						log.Fatal("Could not create .envrc file.", err)
					}
				}
				type Prepend struct {
					Prog                           Program
					Li                             Licence
					Author, Email, Details, Extras string
					Year                           int
				}
				prep := Prepend{
					Prog: Program{
						Name: program,
						Desc: description,
					},
					Li: Licence{
						Short: licences[licence].Short,
					},
					Details: details,
					Year:    year,
					Author:  fullname,
					Extras:  extraText,
					Email:   email,
				}
				// Prepend information.
				prepTmpl, err := template.New("prepend").
					Parse(`
		{{.Prog.Name}}: {{.Prog.Desc}}

		{{.Li.Short}}
		{{.Details}}

		Copyright ©{{.Year}} {{.Author}} all rights reserved.
		Inquiries may be directed to this address: {{.email}}.
		{{.Extras}}
						`)
				if err != nil {
					panic("could not parse template.")
				}
				// create a writer for me to write this template to.
				var prependContent strings.Builder
				err = prepTmpl.Execute(&prependContent, prep)
				if err != nil {
					log.Fatal("Could not execute template.", err)
				}
				// add the writer to the the envrc
				// create a new template for the rest of the envrc
				type envrcEnvironment struct {
					Prep      Prepend
					Variables map[string]any
				}

				// Handling the environment variables for .envrc
				vars := map[string]string{
					"":     "value1",
					"VAR2": "value2",
					// Add your other environment variables here
				}

				type EnvVars struct {
					Variables map[string]string
				}

				envData := EnvVars{Variables: vars}

				envTemplate := `
			{{- range $key, $value := .Variables }}
			export {{ $key }}={{ $value }}
			{{- end }}
			`

				envTmpl, err := template.New("env").Parse(envTemplate)
				if err != nil {
					log.Fatal("Could not parse env template.", err)
				}

				var envContent strings.Builder
				err = envTmpl.Execute(&envContent, envData)
				if err != nil {
					log.Fatal("Could not execute env template.", err)
				}

				// Write the environment variables to the .envrc file
				_, err = file.WriteString(envContent.String())
				if err != nil {
					log.Fatal("Could not write env variables to .envrc file.", err)
				} // loop through all the variables in the template, placing before each key and value of variables, the word export and between them the .
				defer file.Close()
				envrcBytes = append([]byte(envrc), envrcBytes...)
				file.Write(envrcBytes)
			}
			// TODO: create the envrc file and depending on the prepend, alias the touch command, or run befoss add to add the info the files being tracked with git, you can also make this a githook.
			// TODO: ignore all files that git does and make it configurable.
			// after finishing the form let's launch the confetti command.
			if exec.Command("confetti").Run(); err != nil {
				log.Info("No confetti, for more fun please install the confetti commmand. For more info go [here](https://github.com/masalani/confetti)")
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
