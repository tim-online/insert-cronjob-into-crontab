package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/tim-online/insert-cronjob-into-crontab/logger"
)

// Initialize cronjob parser
var (
	parser           = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	assignmentRegexp = regexp.MustCompile(`^\s*\w*\s*=\s*\w*`)
)

type App struct {
	log    *logger.Logger
	Config Config
}

func NewApp() *App {
	app := &App{}

	logger, err := app.newLogger()
	if err != nil {
		log.Fatal(err)
	}

	err = app.initializeConfig()
	if err != nil {
		log.Fatal(err)
	}

	// initialize logger again but with config value
	logger, err = app.newLogger()
	if err != nil {
		app.log.Fatal(err)
	}
	app.setLogger(logger)

	return app
}

func (a *App) initializeConfig() error {
	return nil
}

func (a *App) newLogger() (*logger.Logger, error) {
	log := logger.New(logger.Config{})
	return log, nil
}

func (a *App) setLogger(log *logger.Logger) error {
	a.log = log
	return nil
}

func (a *App) Run(alias string, cronLines []string) error {
	for _, cronLine := range cronLines {
		err := checkCronLine(cronLine)
		if err != nil {
			return err
		}
	}
	cronLine := cronLines[0]

	// stdin is "just" a file so stat it
	stdin := os.Stdin
	fi, err := stdin.Stat()
	if err != nil {
		return err
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// no pipe
		return errors.New("Please provide a crontab via stdin")
	}

	// read all stdin
	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return err
	}

	// search for lines matching the comment
	r := bytes.NewReader(b)

	// check if stdin contains a valid crontab
	err = checkCrontab(r)
	r.Seek(0, 0)
	if err != nil {
		return err
	}

	linesWithComment := findLinesWithAlias(alias, r)
	r.Seek(0, 0)
	if len(linesWithComment) > 1 {
		return errors.Errorf(`Found too many lines matching '%s'. Don't know what to do ¯\_(ツ)_/¯`, alias)
	}

	// replace existing line
	if len(linesWithComment) > 0 {
		lineNo := 0
		for l, _ := range linesWithComment {
			lineNo = l
		}
		replaceLine(r, lineNo, cronLine, os.Stdout)
		return nil
	}

	entry := fmt.Sprintf(`# %s
%s`, alias, cronLine)

	io.Copy(os.Stdout, r)
	os.Stdout.Write([]byte(entry))
	os.Stdout.Write([]byte(string('\n')))

	return err
}

func findLinesWithAlias(alias string, crontab io.Reader) map[int]string {
	comment := fmt.Sprintf("# %s", alias)
	scanner := bufio.NewScanner(crontab)
	i := 0
	linesWithComment := map[int]string{}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == comment {
			linesWithComment[i] = line
		}
		i = i + 1
	}

	return linesWithComment
}

func replaceLine(crontab io.Reader, lineNo int, cronLine string, w io.Writer) {
	scanner := bufio.NewScanner(crontab)
	i := 0
	for scanner.Scan() {
		if i == lineNo+1 {
			w.Write([]byte(cronLine))
			w.Write([]byte(string('\n')))
			i = i + 1
			continue
		}

		i = i + 1
		w.Write(scanner.Bytes())
		w.Write([]byte(string('\n')))
	}

	return
}

func checkCronLine(cronLine string) error {
	// Check if cronjob contains enough fields
	pieces := strings.Split(cronLine, " ")
	if len(pieces) < 6 {
		_, err := parser.Parse(cronLine)
		return errors.Errorf("%s is not a valid cron expression (%s)", cronLine, err)
	}

	// Check if cronjob is valid
	spec := strings.Join(pieces[0:5], " ")
	// cmd := strings.Join(pieces[5:], " ")
	_, err := parser.Parse(spec)
	if err != nil {
		return errors.Errorf("%s is not a valid cron expression (%s)", spec, err)
	}

	return nil
}

func checkCrontab(crontab io.Reader) error {
	scanner := bufio.NewScanner(crontab)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			i = i + 1
			continue
		}

		if lineIsComment(line) {
			i = i + 1
			continue
		}

		if lineIsAssignment(line) {
			i = i + 1
			continue
		}

		// not a comment & not an assignment: should be a valid cron expression
		err := checkCronLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func lineIsComment(line string) bool {
	return strings.HasPrefix(line, "#")
}

func lineIsAssignment(line string) bool {
	return assignmentRegexp.MatchString(line)
}
