package ftpconnection_test

const (
	uid uint = 1

	remotePath = "/foo/bar/baz"

	host     = "ftp-dev-client"
	user     = "user01"
	password = "pwd01"

	// dateFormat = "2006-01-02 15:04"
	//
	// lastModificationDateFormat = "Jan 2 15:04"
)

// Feature messages returned by FEAT command
const (
	featureMsg = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 UTF8
 MLST
211 End`

	featureMsgWithoutMLST = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 UTF8
211 End`

	featureMsgWithoutUTF8 = `211-Features:
 EPRT
 EPSV
 MDTM
 PASV
 REST STREAM
 SIZE
 TVFS
 MLST
211 End`
)

// Status messages returned by STAT and SYST commands.
const (
	statusMsg = `211-FTP server status:
     Connected to 172.22.0.2
     Logged in as ftpuser01
     TYPE: BINARY
     No session bandwidth limit
     Session timeout in seconds is 300
     Control connection is plain text
     Data connections will be plain text
     At session startup, client count was 1
     vsFTPd 3.0.2 - secure, fast, stable
211 End of status`

	systemMsg = `UNIX Type: L8`
)

//const (
//	extendedPassiveModeMessage = "Entering Extended Passive Mode (|||21103|)."
//)
//
//func getEntriesAsListMessage(t *testing.T) string {
//	var builder strings.Builder
//	for _, entry := range getEntries(t) {
//		builder.WriteString(fmt.Sprintf(
//			"%s%s %5d %10s %10s %20d %s %s",
//			entryTypeToStr(t, entry.Type),
//			entry.Permissions,
//			entry.NumHardLinks,
//			entry.OwnerUser,
//			entry.OwnerGroup,
//			entry.SizeInBytes,
//			entry.LastModificationDate.Format(lastModificationDateFormat),
//			entry.Name,
//		))
//	}
//	return builder.String()
//}
//
//func getEntries(t *testing.T) []*entities.Entry {
//	return []*entities.Entry{
//		newEntry(t, "file5", 167, "2022-01-12 16:23"),
//		newEntry(t, "file3", 40032, "2022-01-24 16:23"),
//		newEntry(t, "file9", 2, "2022-01-02 11:23"),
//		newEntry(t, "file2", 102, "2021-05-02 17:23"),
//		newEntry(t, "file6", 9635, "2022-01-02 13:23"),
//		newEntry(t, "file7", 4352, "2020-04-02 14:23"),
//		newEntry(t, "file1", 100, "2022-01-02 15:23"),
//		newEntry(t, "file8", 1034, "2022-09-02 10:23"),
//		newEntry(t, "file4", 5043, "2022-01-12 19:23"),
//	}
//}
//
//func newEntry(t *testing.T, name string, sizeInBytes uint64, dateStr string) *entities.Entry {
//	date, err := time.Parse(dateFormat, dateStr)
//	require.NoError(t, err, "Failed to parse test case date")
//
//	return &entities.Entry{
//		Type:                 entities.EntryTypeFile,
//		Permissions:          "rwxrwxrwx",
//		Name:                 name,
//		OwnerUser:            "user01",
//		OwnerGroup:           "group01",
//		SizeInBytes:          sizeInBytes,
//		NumHardLinks:         2,
//		LastModificationDate: date,
//	}
//}
//
//func entryTypeToStr(t *testing.T, entryType entities.EntryType) string {
//	switch entryType {
//	case entities.EntryTypeFile:
//		return "f"
//	case entities.EntryTypeDir:
//		return "d"
//	case entities.EntryTypeLink:
//		return "l"
//	default:
//		require.Fail(t, "unknown entry type %d", entryType)
//	}
//	return ""
//}
