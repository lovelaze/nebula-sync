package leases

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := writer.Create(name)
		require.NoError(t, err)
		_, err = entry.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	return buf.Bytes()
}

func readArchive(t *testing.T, archive []byte) map[string]string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)

	files := map[string]string{}
	for _, file := range reader.File {
		content, err := file.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(content)
		require.NoError(t, err)
		require.NoError(t, content.Close())
		files[file.Name] = string(data)
	}

	return files
}

func TestFromTeleporter(t *testing.T) {
	archive := newArchive(t, map[string]string{
		"etc/pihole/pihole.toml": "toml",
		TeleporterPath:           "0 aa:aa:aa:aa:aa:01 192.168.1.10 laptop *\n",
	})

	content, err := FromTeleporter(archive)

	require.NoError(t, err)
	assert.Equal(t, "0 aa:aa:aa:aa:aa:01 192.168.1.10 laptop *\n", string(content))
}

func TestFromTeleporter_NoLeases(t *testing.T) {
	content, err := FromTeleporter(newArchive(t, map[string]string{"etc/pihole/pihole.toml": "toml"}))

	require.NoError(t, err)
	assert.Nil(t, content)
}

func TestFromTeleporter_InvalidArchive(t *testing.T) {
	_, err := FromTeleporter([]byte("not a zip"))

	require.Error(t, err)
}

func TestReplaceInTeleporter(t *testing.T) {
	archive := newArchive(t, map[string]string{
		"etc/pihole/pihole.toml": "toml",
		"etc/pihole/gravity.db":  "db",
		TeleporterPath:           "old\n",
	})

	replaced, err := ReplaceInTeleporter(archive, []byte("new\n"))

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"etc/pihole/pihole.toml": "toml",
		"etc/pihole/gravity.db":  "db",
		TeleporterPath:           "new\n",
	}, readArchive(t, replaced))
}

func TestReplaceInTeleporter_AddsMissingLeases(t *testing.T) {
	archive := newArchive(t, map[string]string{"etc/pihole/pihole.toml": "toml"})

	replaced, err := ReplaceInTeleporter(archive, []byte("new\n"))

	require.NoError(t, err)
	assert.Equal(t, "new\n", readArchive(t, replaced)[TeleporterPath])
}

func TestNewTeleporter(t *testing.T) {
	archive, err := NewTeleporter([]byte("leases\n"))

	require.NoError(t, err)
	assert.Equal(t, map[string]string{TeleporterPath: "leases\n"}, readArchive(t, archive))

	empty, err := NewTeleporter(nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{TeleporterPath: ""}, readArchive(t, empty))
}
