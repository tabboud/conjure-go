// Copyright (c) 2018 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conjure

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/palantir/go-ptimports/ptimports"
	"github.com/palantir/goastwriter"
	"github.com/palantir/goastwriter/astgen"
	"github.com/pkg/errors"
)

type OutputFile struct {
	absPath    string
	pkgName    string
	goTypeObjs []astgen.ASTDecl
}

func (f *OutputFile) AbsPath() string {
	return f.absPath
}

func (f *OutputFile) Write() error {
	body, err := f.Render()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.absPath), 0755); err != nil {
		return errors.Wrapf(err, "failed to create parent directory for Go file output %s", f.absPath)
	}
	if err := ioutil.WriteFile(f.absPath, body, 0644); err != nil {
		return errors.Wrapf(err, "failed to write Go file output to %s", f.absPath)
	}
	return nil
}

func (f *OutputFile) Render() ([]byte, error) {
	goFileSrc, err := goastwriter.Write(f.pkgName, f.goTypeObjs...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate Go source for file %s", f.absPath)
	}
	goFileSrc, err = ptimports.Process("", goFileSrc, &ptimports.Options{Refactor: true})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run ptimports on generated Go source for file %s", f.absPath)
	}
	// add extra newline after braces
	goFileSrc = regexp.MustCompile(`(\n}\n)(\S)`).ReplaceAll(goFileSrc, []byte("$1\n$2"))
	goFileSrc = addHeaderComment(goFileSrc)

	return goFileSrc, nil
}

func addHeaderComment(bytes []byte) []byte {
	return append([]byte(`// This file was generated by Conjure and should not be manually edited.

`), bytes...)
}
