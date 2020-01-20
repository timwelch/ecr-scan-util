package helpers

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/google/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type ReporterConfig struct {
	ReportFileName string
	ReporterType   string
	ReportBaseDir  string
}

func NewDefaultReporterConfig() (config ReporterConfig) {
	return ReporterConfig{
		ReportFileName: "testreport.xml",
		ReporterType:   "junit",
		ReportBaseDir:  "reports",
	}
}

func NewCustomReporterConfig(filename string, basedir string, reporterType string) (config ReporterConfig) {
	return ReporterConfig{
		ReportFileName: filename,
		ReportBaseDir:  basedir,
		ReporterType:   reporterType,
	}
}

type GlobalConfig struct {
	AwsConfig      *aws.Config
	ReporterConfig ReporterConfig
}

func Check(e error, logger logger.Logger, a ...interface{}) {
	if e != nil {
		logger.Error(a)
		panic(e)
	}
}
func CheckAndExit(e error, logger logger.Logger, a ...interface{}) {
	if e != nil {
		logger.Fatal(a)
		os.Exit(1)
	}
}
func CompositionParser(compositionFile string, stripPrefix string, stripSuffix string, l logger.Logger) (map[string]string, error) {
	zdComposition := make(map[string]string)
	containerList := make(map[string]string)

	l.Infof("trying to read container names and identifiers from %s", compositionFile)
	yamlFile, err := ioutil.ReadFile(compositionFile)
	Check(err, l, "Failed to read file %s: %s", compositionFile, err)

	l.Infof("unmarshalling contents of %s", compositionFile)
	err = yaml.Unmarshal(yamlFile, zdComposition)
	Check(err, l, "Failed to unmarshal %s, %v", yamlFile, err)

	for c, v := range zdComposition {
		c = underscoreHyphenator(suffixStripper(prefixStripper(c, stripPrefix), stripSuffix))
		containerList[c] = v
	}

	return containerList, err
}

func ExtractPackageAttributes(query string, finding *ecr.ImageScanFinding) (attribute string, err error) {
	for a := range finding.Attributes {
		if *finding.Attributes[a].Key == query {
			attribute = *finding.Attributes[a].Value
		}
	}
	if attribute != "" {
		return attribute, nil
	} else {
		fmt.Printf("Query for key %s returned no hits or an emtpy value", query)
		return "", errors.New("query for returned an empty result or key has no associated value")
	}
}

func FileNameFormatter(filename string) string {

	return path.Base(fmt.Sprintf("%s-%s.xml", filename, timeStamper()))
}

func timeStamper() string {
	t := time.Now().Format(time.RFC3339)
	t = strings.Replace(t, "Z", "", 1)
	t = strings.Replace(t, "-", "", -1)
	t = strings.Replace(t, ":", "", -1)
	t = strings.Replace(t, "T", "-", 1)
	return strings.Replace(t, " ", "", -1)
}

func suffixStripper(input string, suffix string) (output string) {

	//We have to do some extra work because some containers we run
	//use version in the middle of their name which we can't strip.

	i := strings.LastIndex(input, suffix)
	if i != -1 {
		input = input[:i] + strings.Replace(input[i:], suffix, "", 1)
	}

	return input
}

func prefixStripper(input string, prefix string) (output string) {

	//prefix stripper doesn't just removes the first instance

	return strings.Replace(input, prefix, "", 1)
}

func underscoreHyphenator(input string) (output string) {
	return strings.Replace(input, "_", "-", -1)
}
