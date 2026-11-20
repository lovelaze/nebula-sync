package leases

import (
	"archive/zip"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"time"
)

// TeleporterPath is where a Teleporter archive keeps the lease database.
const TeleporterPath = "etc/pihole/dhcp.leases"

// leasesFileMode is the mode dnsmasq creates the lease database with.
const leasesFileMode = 0o644

// FromTeleporter returns the lease database of a Teleporter archive, or nil
// if the archive has none (the instance never ran a DHCP server).
func FromTeleporter(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open teleporter archive: %w", err)
	}

	for _, file := range reader.File {
		if file.Name == TeleporterPath {
			return readFile(file)
		}
	}

	return nil, nil
}

func readFile(file *zip.File) ([]byte, error) {
	content, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer content.Close()

	return io.ReadAll(content)
}

// ReplaceInTeleporter returns a copy of a Teleporter archive whose lease
// database is replaced by (or, if it had none, extended with) leases.
func ReplaceInTeleporter(archive, leases []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open teleporter archive: %w", err)
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, file := range reader.File {
		if file.Name == TeleporterPath {
			continue
		}
		if err := writer.Copy(file); err != nil {
			return nil, fmt.Errorf("copy %s: %w", file.Name, err)
		}
	}

	if err := addLeases(writer, leases); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close teleporter archive: %w", err)
	}

	return buf.Bytes(), nil
}

// NewTeleporter returns a Teleporter archive that holds only a lease database.
func NewTeleporter(leases []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	if err := addLeases(writer, leases); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close teleporter archive: %w", err)
	}

	return buf.Bytes(), nil
}

// addLeases stores the lease database uncompressed with its size and checksum
// in the local header, the simplest entry any zip reader handles.
func addLeases(writer *zip.Writer, leases []byte) error {
	header := &zip.FileHeader{
		Name:               TeleporterPath,
		Comment:            "DHCP leases file",
		Method:             zip.Store,
		Modified:           time.Now(),
		CRC32:              crc32.ChecksumIEEE(leases),
		CompressedSize64:   uint64(len(leases)),
		UncompressedSize64: uint64(len(leases)),
	}
	header.SetMode(leasesFileMode)

	entry, err := writer.CreateRaw(header)
	if err != nil {
		return fmt.Errorf("add %s: %w", TeleporterPath, err)
	}
	if _, err := entry.Write(leases); err != nil {
		return fmt.Errorf("write %s: %w", TeleporterPath, err)
	}

	return nil
}
