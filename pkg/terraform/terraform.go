package terraform

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/installer/data"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/openshift/installer/pkg/lineprinter"
	texec "github.com/openshift/installer/pkg/terraform/exec"
)

const (
	// StateFileName is the default name for Terraform state files.
	StateFileName string = "terraform.tfstate"

	// VarFileName is the default name for Terraform var file.
	VarFileName string = "terraform.tfvars"
)

// Version gets the output of 'terrraform version'.
func Version() (version string, err error) {
	return texec.Version(), nil
}

// Apply unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// apply'.  It returns the absolute path of the tfstate file, rooted
// in the specified directory, along with any errors from Terraform.
func Apply(dir string, platform string, extraArgs ...string) (path string, err error) {
	err = unpackAndInit(dir, platform)
	if err != nil {
		return "", err
	}

	defaultArgs := []string{
		"-auto-approve",
		"-input=false",
		fmt.Sprintf("-state=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-state-out=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-var-file=%s", filepath.Join(dir, VarFileName)),
	}
	args := append(defaultArgs, extraArgs...)
	args = append(args, dir)
	sf := filepath.Join(dir, StateFileName)

	tDebug := &lineprinter.Trimmer{WrappedPrint: logrus.Debug}
	tError := &lineprinter.Trimmer{WrappedPrint: logrus.Error}
	lpDebug := &lineprinter.LinePrinter{Print: tDebug.Print}
	lpError := &lineprinter.LinePrinter{Print: tError.Print}
	defer lpDebug.Close()
	defer lpError.Close()

	if exitCode := texec.Apply(dir, args, lpDebug, lpError); exitCode != 0 {
		return sf, errors.New("failed to apply using terraform")
	}
	return sf, nil
}

// Destroy unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// destroy'.
func Destroy(dir string, platform string, extraArgs ...string) (err error) {
	err = unpackAndInit(dir, platform)
	if err != nil {
		return err
	}

	defaultArgs := []string{
		"-auto-approve",
		"-input=false",
		fmt.Sprintf("-state=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-state-out=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-var-file=%s", filepath.Join(dir, VarFileName)),
	}
	args := append(defaultArgs, extraArgs...)
	args = append(args, dir)

	tDebug := &lineprinter.Trimmer{WrappedPrint: logrus.Debug}
	tError := &lineprinter.Trimmer{WrappedPrint: logrus.Error}
	lpDebug := &lineprinter.LinePrinter{Print: tDebug.Print}
	lpError := &lineprinter.LinePrinter{Print: tError.Print}
	defer lpDebug.Close()
	defer lpError.Close()

	if exitCode := texec.Destroy(dir, args, lpDebug, lpError); exitCode != 0 {
		return errors.New("failed to destroy using terraform")
	}
	return nil
}

// unpack unpacks the platform-specific Terraform modules into the
// given directory.
func unpack(dir string, platform string) (err error) {
	err = data.Unpack(dir, platform)
	if err != nil {
		return err
	}

	err = data.Unpack(filepath.Join(dir, "config.tf"), "config.tf")
	if err != nil {
		return err
	}

	return nil
}

// unpackAndInit unpacks the platform-specific Terraform modules into
// the given directory and then runs 'terraform init'.
func unpackAndInit(dir string, platform string) (err error) {
	err = unpack(dir, platform)
	if err != nil {
		return errors.Wrap(err, "failed to unpack Terraform modules")
	}

	tDebug := &lineprinter.Trimmer{WrappedPrint: logrus.Debug}
	tError := &lineprinter.Trimmer{WrappedPrint: logrus.Error}
	lpDebug := &lineprinter.LinePrinter{Print: tDebug.Print}
	lpError := &lineprinter.LinePrinter{Print: tError.Print}
	defer lpDebug.Close()
	defer lpError.Close()

	if exitCode := texec.Init(dir, []string{dir}, lpDebug, lpError); exitCode != 0 {
		return errors.New("failed to initialize Terraform")
	}
	return nil
}
