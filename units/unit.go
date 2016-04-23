package units

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/ocitools/config"
)

const (
	testCacheDir  = "./bundles/"
	runtimePrefix = "runtime.json"
	configPrefix  = "config.json"
	pass          = "SUCCESS"
	fail          = "FAILED"
)

// TestUnit for storage testcase
type TestUnit struct {
	Runtime string

	// Case name
	Name string
	// Args is used to generate bundle
	Args string
	// Describle what does this unit test for. It is optional.
	Description string

	BundleDir string
	// Success or failed
	Result string
	// When result == failed, ErrInfo is err code, or, ErrInfo is nil
	ErrInfo error
}

type testUnits []*TestUnit

// Units is the object of testUnits
var Units = new(testUnits)

// LoadTestUnits load TestUnits configuration from config
func LoadTestUnits(filename string) {

	for key, value := range config.BundleMap {
		// TODO: config.BundleMap should support 'Description'
		unit := NewTestUnit(key, value, "")
		*Units = append(*Units, unit)
	}
}

// NewTestUnit new a TestUnit
func NewTestUnit(name string, args string, desc string) *TestUnit {
	tu := new(TestUnit)
	tu.Name = name
	tu.Args = args
	tu.Description = desc

	return tu
}

// OutputResult output results, ouput value: err-only or all
func OutputResult(output string) {
	if output != "err-only" && output != "all" {
		logrus.Fatalf("eerror output mode: %v\n", output)
	}

	SuccessCount := 0
	failCount := 0

	// Can not be merged into on range, because output should be devided into two parts, successful and
	// failure
	if output == "all" {
		logrus.Println("successful Details:")
		echoDividing()
	}

	for _, tu := range *Units {
		if tu.Result == pass {
			SuccessCount++
			if output == "all" {
				tu.EchoSUnit()
			}
		}
	}

	logrus.Println("failure Details:")
	echoDividing()

	for _, tu := range *Units {
		if tu.Result == fail {
			failCount++
			tu.EchoFUnit()
		}
	}

	echoDividing()
	logrus.Printf("statistics:  %v bundles success, %v bundles failed\n", SuccessCount, failCount)
}

// EchoSUnit echo sucessful test units after validation
func (unit *TestUnit) EchoSUnit() {
	logrus.Printf("\nBundleName:\n  %v\nBundleDir:\n  %v\nCaseArgs:\n  %v\nTestResult:\n  %v\n",
		unit.Name, unit.BundleDir, unit.Args, unit.Result)
}

// EchoFUnit echo failed test units after validation
func (unit *TestUnit) EchoFUnit() {
	logrus.Printf("\nBundleName:\n  %v\nBundleDir:\n  %v\nCaseArgs:\n  %v\nResult:\n  %v\n"+
		"ErrInfo:\n  %v\n", unit.Name, unit.BundleDir, unit.Args, unit.Result, unit.ErrInfo)
}

func echoDividing() {
	logrus.Println("============================================================================" +
		"===================")
}

func (unit *TestUnit) setResult(result string, err error) {
	unit.Result = result
	if result == pass {
		unit.ErrInfo = nil
	} else {
		unit.ErrInfo = err
	}
}

// SetRuntime set runtime for validation
func (unit *TestUnit) SetRuntime(runtime string) error {
	unit.Runtime = "runc"
	return nil
}

// Run run testunits
func (unit *TestUnit) Run() {
	if unit.Runtime == "" {
		logrus.Fatalf("set the runtime before testing")
	} else if unit.Runtime != "runc" {
		logrus.Fatalf("%v have not supported yet\n", unit.Runtime)
	}

	unit.generateConfigs()
	unit.prepareBundle()

	if _, err := runcStart(unit.BundleDir); err != nil {
		unit.setResult(fail, err)
		return
	}

	unit.setResult(pass, nil)
	return
}

func (unit *TestUnit) prepareBundle() {
	// Create bundle follder
	unit.BundleDir = path.Join(testCacheDir, unit.Name)
	if err := os.RemoveAll(unit.BundleDir); err != nil {
		logrus.Fatalf("remove bundle %v err: %v\n", unit.Name, err)
	}

	if err := os.Mkdir(unit.BundleDir, os.ModePerm); err != nil {
		logrus.Fatalf("mkdir bundle %v dir err: %v\n", unit.BundleDir, err)
	}

	// Create rootfs for bundle
	rootfs := unit.BundleDir + "/rootfs"
	if err := untarRootfs(rootfs); err != nil {
		logrus.Fatalf("tar roofts.tar.gz to %v err: %v\n", rootfs, err)
	}

	// Copy runtimtest from plugins to rootfs
	src := "./runtimetest"
	dRuntimeTest := rootfs + "/runtimetest"

	if err := copy(dRuntimeTest, src); err != nil {
		logrus.Fatalf("Copy runtimetest to rootfs err: %v\n", err)
	}

	if err := os.Chmod(dRuntimeTest, os.ModePerm); err != nil {
		logrus.Fatalf("Chmod runtimetest mode err: %v\n", err)
	}

	// Copy *.json to testroot and rootfs
	csrc := configPrefix
	rsrc := runtimePrefix
	cdest := rootfs + "/" + configPrefix
	rdest := rootfs + "/" + runtimePrefix

	if err := copy(cdest, csrc); err != nil {
		logrus.Fatal(err)
	}

	if err := copy(rdest, rsrc); err != nil {
		logrus.Fatal(err)
	}

	cdest = unit.BundleDir + "/" + configPrefix
	rdest = unit.BundleDir + "/" + runtimePrefix

	if err := copy(cdest, csrc); err != nil {
		logrus.Fatal(err)
	}

	if err := copy(rdest, rsrc); err != nil {
		logrus.Fatal(err)
	}
}

func (unit *TestUnit) generateConfigs() {
	args := splitArgs(unit.Args)

	logrus.Debugf("args to the ocitools generate: ")
	for _, a := range args {
		logrus.Debugln(a)
	}

	err := genConfigs(args)
	if err != nil {
		logrus.Fatalf("generate *.json err: %v\n", err)
	}
}

func untarRootfs(rootfs string) error {
	// Create rootfs folder to bundle
	if err := os.Mkdir(rootfs, os.ModePerm); err != nil {
		logrus.Fatalf("mkdir rootfs for bundle %v err: %v\n", rootfs, err)
	}

	cmd := exec.Command("tar", "-xf", "rootfs.tar.gz", "-C", rootfs)
	cmd.Dir = ""
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()

	logrus.Debugln("command done\n")
	logrus.Debugln(string(out))
	if err != nil {
		return err
	}
	return nil
}

func genConfigs(args []string) error {
	argsNew := make([]string, len(args)+1)
	argsNew[0] = "generate"
	for i, a := range args {
		argsNew[i+1] = a
	}

	cmd := exec.Command("./ocitools", argsNew...)
	cmd.Dir = "./"
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	logrus.Debugf("command done\n")
	logrus.Debugln(string(out))
	if err != nil {
		return err
	}
	return nil
}

func runcStart(specDir string) (string, error) {
	logrus.Debugf("launcing runtime")

	cmd := exec.Command("runc", "start")
	cmd.Dir = specDir
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()

	logrus.Debugf("command done")
	if err != nil {
		return string(out), errors.New(string(out) + err.Error())
	}
	return string(out), nil
}

func splitArgs(args string) []string {
	argsnew := strings.TrimSpace(args)
	argArray := strings.Split(argsnew, "--")

	length := len(argArray)
	resArray := make([]string, length-1)
	for i, arg := range argArray {
		if i == 0 || i == length {
			continue
		} else {
			resArray[i-1] = "--" + strings.TrimSpace(arg)
		}
	}
	return resArray
}

func copy(dst string, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}
