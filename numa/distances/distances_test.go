/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package distances

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	fakesysfs "github.com/fromanirh/topologyinfo/sysfs/fake"
)

type distTcase struct {
	Name string
	From int
	To   int
}

func TestDistancesBetweenNodes(t *testing.T) {
	dists, err := NewDistancesFromData(map[string]string{
		"0": "10\n",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	val, err := dists.BetweenNodes(0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != 10 {
		t.Errorf("unexpected distance: got %v expected %v", val, 10)
	}

	for _, tc := range []distTcase{
		{"unknown to", 0, 1},
		{"unknown from", 1, 0},
		{"negative to", 0, -1},
		{"negative from", -1, 0},
		{"both unknown", 255, 255},
		{"both negative", -1, -1},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			if _, err = dists.BetweenNodes(tc.From, tc.To); err == nil {
				t.Errorf("got distance between unknown nodes: %d, %d", tc.From, tc.To)
			}

		})
	}

}

type tcase struct {
	from int
	to   int
	dist int
}

func TestReadDistances(t *testing.T) {

	distData := map[string]string{
		"0": "10 21",
		"1": "21 10",
	}

	tcases := []tcase{
		{0, 0, 10},
		{1, 1, 10},
		{0, 1, 21},
		{1, 0, 21},
	}

	nodeData := map[string]string{
		"online":            "0,1",
		"possible":          "0,1",
		"has_cpu":           "0,1",
		"has_memory":        "0,1",
		"has_normal_memory": "0,1",
	}

	base, err := ioutil.TempDir("/tmp", "fakesysfs")
	if err != nil {
		t.Errorf("error creating temp base dir: %v", err)
	}
	fs, err := fakesysfs.NewFakeSysfs(base)
	if err != nil {
		t.Errorf("error creating fakesysfs: %v", err)
	}
	t.Logf("sysfs at %q", fs.Base())

	sysDevs := fs.AddTree("sys", "devices")
	devSys := sysDevs.Add("system", nil)
	devNode := devSys.Add("node", nodeData)
	for nodeID, nodeDists := range distData {
		devNode.Add(fmt.Sprintf("node%s", nodeID), map[string]string{
			"distance": fmt.Sprintf("%s\n", nodeDists),
		})
	}

	err = fs.Setup()
	if err != nil {
		t.Errorf("error setting up fakesysfs: %v", err)
	}
	defer func() {
		if _, ok := os.LookupEnv("FAKESYSFS_KEEP_TREE"); ok {
			t.Logf("found environment variable, keeping fake tree")
		} else {
			err = fs.Teardown()
			if err != nil {
				t.Errorf("error tearing down fakesysfs: %v", err)
			}
		}
	}()

	dists, err := NewDistancesFromSysfs(filepath.Join(fs.Base(), "sys"))
	if err != nil {
		t.Errorf("error in NewDistancesFromSysFS: %v", err)
	}

	for _, tc := range tcases {
		t.Run(fmt.Sprintf("%d->%d", tc.from, tc.to), func(t *testing.T) {
			val, err := dists.BetweenNodes(tc.from, tc.to)
			if err != nil {
				t.Errorf("error in BetweenNodes(%d, %d): %v", tc.from, tc.to, err)
			}
			if val != tc.dist {
				t.Errorf("distance mismatch found %v expected %v", val, tc.dist)
			}
		})
	}
}
