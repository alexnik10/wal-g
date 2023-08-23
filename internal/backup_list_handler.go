package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/jedib0t/go-pretty/table"
	"github.com/wal-g/tracelog"
	"github.com/wal-g/wal-g/pkg/storages/storage"
)

func HandleDefaultBackupList(folder storage.Folder, pretty, json bool) {
	getBackupsFunc := func() ([]BackupTime, error) {
		res, err := GetBackups(folder)
		if _, ok := err.(NoBackupsFoundError); ok {
			err = nil
		}
		return res, err
	}
	writeBackupListFunc := func(backups []BackupTime) {
		SortBackupTimeSlices(backups)
		switch {
		case json:
			err := WriteAsJSON(backups, os.Stdout, pretty)
			tracelog.ErrorLogger.FatalOnError(err)
		case pretty:
			WritePrettyBackupList(backups, os.Stdout)
		default:
			WriteBackupList(backups, os.Stdout)
		}
	}

	HandleBackupList(getBackupsFunc, writeBackupListFunc)
}

func HandleBackupList(
	getBackupsFunc func() ([]BackupTime, error),
	writeBackupListFunc func([]BackupTime),
) {
	backups, err := getBackupsFunc()
	tracelog.ErrorLogger.FatalOnError(err)

	if len(backups) == 0 {
		tracelog.InfoLogger.Println("No backups found")
		return
	}

	writeBackupListFunc(backups)
}

func WriteBackupList(backups []BackupTime, output io.Writer) {
	writer := tabwriter.NewWriter(output, 0, 0, 1, ' ', 0)
	defer writer.Flush()
	fmt.Fprintln(writer, "name\tmodified\twal_segment_backup_start")
	for _, b := range backups {
		fmt.Fprintf(writer, "%v\t%v\t%v\n", b.BackupName, FormatTime(b.Time), b.WalFileName)
	}
}

func WritePrettyBackupList(backups []BackupTime, output io.Writer) {
	writer := table.NewWriter()
	writer.SetOutputMirror(output)
	defer writer.Render()
	writer.AppendHeader(table.Row{"#", "Name", "Modified", "WAL segment backup start"})
	for i, b := range backups {
		writer.AppendRow(table.Row{i, b.BackupName, PrettyFormatTime(b.Time), b.WalFileName})
	}
}

func WriteAsJSON(data interface{}, output io.Writer, pretty bool) error {
	var bytes []byte
	var err error
	if pretty {
		bytes, err = json.MarshalIndent(data, "", "    ")
	} else {
		bytes, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}
	_, err = output.Write(bytes)
	return err
}
