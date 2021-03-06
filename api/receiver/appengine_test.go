// +build medium

// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/web-platform-tests/wpt.fyi/shared"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
)

type mockGcs struct {
	mockWriter mockGcsWriter
}

// mockGcsWriter implements io.WriteCloser
type mockGcsWriter struct {
	bytes.Buffer
	finalContent []byte
}

func (m *mockGcsWriter) Close() error {
	m.finalContent = m.Bytes()
	return nil
}

func (m *mockGcs) NewWriter(bucketName, fileName, contentType, contentEncoding string) (io.WriteCloser, error) {
	return &m.mockWriter, nil
}

func TestUploadToGCS(t *testing.T) {
	a := appEngineAPIImpl{}
	mGcs := mockGcs{}
	a.gcs = &mGcs

	buffer := bytes.NewBufferString("test content")
	path, err := a.uploadToGCS("test.json", buffer, false)
	assert.Nil(t, err)

	assert.Equal(t, path, fmt.Sprintf("/%s/test.json", BufferBucket))
	assert.Equal(t, string(mGcs.mockWriter.finalContent), "test content", 0)
}

func TestScheduleResultsTask(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.Nil(t, err)
	defer done()

	stats, err := taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 0)

	a := appEngineAPIImpl{ctx: ctx}
	_, err = a.scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "single", nil)
	assert.Nil(t, err)

	stats, err = taskqueue.QueueStats(ctx, []string{""})
	assert.Nil(t, err)
	assert.Equal(t, stats[0].Tasks, 1)
}

func TestScheduleResultsTask_error(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.Nil(t, err)
	defer done()
	a := appEngineAPIImpl{ctx: ctx}

	_, err = a.scheduleResultsTask("", []string{"/blade-runner/test.json"}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", []string{""}, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", nil, "single", nil)
	assert.NotNil(t, err)

	_, err = a.scheduleResultsTask("blade-runner", []string{"/blade-runner/test.json"}, "", nil)
	assert.NotNil(t, err)
}

func TestAuthenticateUploader(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	assert.Nil(t, err)
	r, err := i.NewRequest("POST", "/api/admin/upload", nil)
	assert.Nil(t, err)
	ctx := appengine.NewContext(r)

	assert.Nil(t, err)
	defer i.Close()

	a := appEngineAPIImpl{ctx: ctx}
	assert.False(t, a.AuthenticateUploader("user", "123"))

	key := datastore.NewKey(ctx, "Uploader", "user", 0, nil)
	datastore.Put(ctx, key, &shared.Uploader{Username: "user", Password: "123"})
	assert.True(t, a.AuthenticateUploader("user", "123"))
}
