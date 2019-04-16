// Copyright (c) 2018-2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package cache

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	client "github.com/sylabs/singularity/pkg/client/library"
)

func TestLibrary(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected string
	}{
		{"Default Library", "", filepath.Join(cacheDefault, "library")},
		{"Custom Library", cacheCustom, filepath.Join(cacheCustom, "library")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer Clean()
			defer os.Unsetenv(DirEnv)

			os.Setenv(DirEnv, tt.env)

			if r := Library(); r != tt.expected {
				t.Errorf("Unexpected result: %s (expected %s)", r, tt.expected)
			}
		})
	}
}

func TestLibraryImage(t *testing.T) {
	LibraryImage("", "")
}

func TestLibraryImageExists(t *testing.T) {
	// Invalid cases
	_, err := LibraryImageExists("", "")
	if err == nil {
		t.Fatalf("LibraryImageExists() returned true for invalid data:  %s\n", err)
	}

	// Pull an image so we know for sure that it is in the cache
	sexec, err := exec.LookPath("singularity")
	if err != nil {
		t.Fatalf("cannot get path for singularity: %s\n", err)
	}
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("cannot create temporary directory: %s\n", err)
	}
	filename := "ubuntu_latest.sif"
	name := dir + filename
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(sexec, "pull", "-F", "-U", name, "library://ubuntu")
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		t.Fatalf("command failed: %s - stdout: %s - stderr: %s\n", err, stdout.String(), stderr.String())
	}
	defer os.RemoveAll(dir)

	// Invalid case with a valid image
	_, err = LibraryImageExists("", filename)
	if err != nil {
		t.Fatalf("image reported as non-existing: %s\n", err)
	}

	// Valid case with a valid image, the get the hash from the
	// file we just created and check whether it matches with what
	// we have in the cache
	hash, err := client.ImageHash(name)
	if err != nil {
		t.Fatalf("cannot get image's hash: %s\n", err)
	}

	exists, err := LibraryImageExists(hash, filename)
	if err != nil {
		t.Fatalf("error while checking if image exists: %s\n", err)
	}
	if exists == false {
		t.Fatal("valid image is reported as non-existing")
	}
}
