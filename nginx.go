package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

var module_requirement = map[string]string{
	"mirror": "1.13",
}

const (
	binary = "/usr/local/openresty/bin/openresty"
	config = "/etc/nginx/nginx.conf"

)

func NginxMirrorModuleCheck() bool {
	fmt.Println("Check nginx mirror module support")
	out, err := exec.Command(binary, "-v").CombinedOutput()
	if err != nil {
		fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return false
	}

	r, _ := regexp.Compile(module_requirement["mirror"])
	if r.MatchString(string(out)) {
		fmt.Println(`
-------------------------------------------------------------------------------
nginx mirror support: true
-------------------------------------------------------------------------------
`)
		return true
	} else {
		fmt.Println(`
-------------------------------------------------------------------------------
nginx mirror support: false
-------------------------------------------------------------------------------
`)
		return false
	}
}

func NginxConfigTestOnly() error {
	out, err := exec.Command(binary, "-t", "-c", config).CombinedOutput()
	if err != nil {
		// this error is different from the rest because it must be clear why nginx is not working
		oe := fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return errors.New(oe)
	}
	return err
}

func NgixnReloadOnly() error {
	out, err := exec.Command(binary, "-s", "reload", "-c", config).CombinedOutput()
	if err != nil {
		// this error is different from the rest because it must be clear why nginx is not working
		oe := fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return errors.New(oe)
	}
	return err
}

func NginxConfigTest(cfg []byte) error {
	tmpfile, err := ioutil.TempFile("", "nginx-cfg")
	if err != nil {
		return err
	}
	defer tmpfile.Close()
	err = ioutil.WriteFile(tmpfile.Name(), cfg, 0644)
	if err != nil {
		return err
	}
	out, err := exec.Command(binary, "-t", "-c", tmpfile.Name()).CombinedOutput()
	if err != nil {
		// this error is different from the rest because it must be clear why nginx is not working
		oe := fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return errors.New(oe)
	}

	os.Remove(tmpfile.Name())
	return nil
}

func NginxReload(cfg []byte) error {
	err := ioutil.WriteFile(config, cfg, 0644)
	if err != nil {
		return err
	}

	out, err := exec.Command(binary, "-s", "reload", "-c", config).CombinedOutput()
	if err != nil {
		// this error is different from the rest because it must be clear why nginx is not working
		oe := fmt.Sprintf(`
-------------------------------------------------------------------------------
Error: %v
%v
-------------------------------------------------------------------------------
`, err, string(out))
		return errors.New(oe)
	}

	return nil
}
