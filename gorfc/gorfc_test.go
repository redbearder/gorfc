package gorfc

import (
	"fmt"
	"os"
	"reflect"

	//"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// NW RFC Lib Version
//
func TestNWRFCLibVersion(t *testing.T) {
	major, minor, patchlevel := GetNWRFCLibVersion()
	assert.Equal(t, uint(7500), major) // adapt to your NW RFC Lib version
	assert.Equal(t, uint(0), minor)
	assert.Greater(t, patchlevel, uint(4))
}

//
// Connection Tests
//
func TestConnect(t *testing.T) {
	fmt.Println("Connection test: Open and Close")
	c, err := ConnectionFromParams(abapSystem())
	if err != nil {
		t.SkipNow()
	}
	assert.NotNil(t, c)
	assert.Nil(t, err)
	assert.True(t, c.Alive())
	c.Close()
	assert.False(t, c.Alive())
}

func TestConnectionAttributes(t *testing.T) {
	fmt.Println("Connection test: Attributes")
	c, err := ConnectionFromParams(abapSystem())
	assert.Equal(t, err, nil)

	a, err := c.GetConnectionAttributes()
	paramNames := map[string]struct{}{
		"Dest":                  struct{}{},
		"Host":                  struct{}{},
		"PartnerHost":           struct{}{},
		"SysNumber":             struct{}{},
		"SysId":                 struct{}{},
		"Client":                struct{}{},
		"User":                  struct{}{},
		"Language":              struct{}{},
		"Trace":                 struct{}{},
		"IsoLanguage":           struct{}{},
		"Codepage":              struct{}{},
		"PartnerCodepage":       struct{}{},
		"RfcRole":               struct{}{},
		"Type":                  struct{}{},
		"PartnerType":           struct{}{},
		"Rel":                   struct{}{},
		"PartnerRel":            struct{}{},
		"KernelRel":             struct{}{},
		"CpicConvId":            struct{}{},
		"ProgName":              struct{}{},
		"PartnerBytesPerChar":   struct{}{},
		"PartnerSystemCodepage": struct{}{},
		"partnerIP":             struct{}{},
		"partnerIPv6":           struct{}{},
	}

	// check if all parameters returned
	assert.Equal(t, len(a), len(paramNames))
	// and the content of some
	assert.Equal(t, strings.ToUpper(abapSystem()["user"]), a["user"])
	assert.Equal(t, abapSystem()["sysnr"], a["sysNumber"])
	assert.Equal(t, abapSystem()["client"], a["client"])
	c.Close()
}

func TestPing(t *testing.T) {
	fmt.Println("Connection test: Ping")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	err = c.Ping()
	assert.Nil(t, err)
	c.Close()
}

func TestReopen(t *testing.T) {
	fmt.Println("Connection test: Reopen")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	err = c.Reopen()
	assert.Nil(t, err)
	c.Close()
}

func TestConnectFromDest(t *testing.T) {
	fmt.Println("Connection test: Destination")
	assert.Greater(t, len(os.Getenv("RFC_INI")), 0)
	c, err := ConnectionFromDest("MME")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	c.Close()
}

func TestConnectionEcho(t *testing.T) {
	fmt.Println("connection test: Echo")
	assert.Greater(t, len(os.Getenv("RFC_INI")), 0)
	c, err := ConnectionFromDest("MME")
	assert.Nil(t, err)
	assert.NotNil(t, c)
	type importStruct struct {
		XXX string
	}
	params := map[string]interface{}{
		"REQUTEXT": "Hällö",
	}
	r, err := c.Call("STFC_CONNECTION", params)
	assert.Nil(t, err)
	assert.NotNil(t, r["ECHOTEXT"])
	assert.Equal(t, params["REQUTEXT"], r["ECHOTEXT"])
	c.Close()
}

//
// Connection Errors
//

func TestWrongUserConnect(t *testing.T) {
	fmt.Println("Connection Error: Logon")
	a := abapSystem()
	a["user"] = "@!n0user"
	c, err := ConnectionFromParams(a)
	assert.Nil(t, c)
	assert.NotNil(t, err)
	assert.Equal(t, "Connection could not be opened", err.(*RfcError).Description)
	assert.Equal(t, "Name or password is incorrect (repeat logon)", err.(*RfcError).ErrorInfo.Message)
	assert.Equal(t, "RFC_LOGON_FAILURE", err.(*RfcError).ErrorInfo.Code)
	assert.Equal(t, "RFC_LOGON_FAILURE", err.(*RfcError).ErrorInfo.Key)
}

func TestMissingAshostConnect(t *testing.T) {
	fmt.Println("Connection Error: Connection parameter missing")
	a := abapSystem()
	a["ashost"] = ""
	c, err := ConnectionFromParams(a)
	assert.Nil(t, c)
	assert.NotNil(t, err)
	assert.Equal(t, "Connection could not be opened", err.(*RfcError).Description)
	assert.Equal(t, "Parameter ASHOST, GWHOST, MSHOST or PORT is missing.", err.(*RfcError).ErrorInfo.Message)
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Code)
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Key)
}

func TestWrongParameter(t *testing.T) {
	fmt.Println("Connection Error: Invoke with wrong parameter")
	type importStruct struct {
		XXX string
	}
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	r, err := c.Call("STFC_CONNECTION", importStruct{"wrong param"})
	assert.Equal(t, map[string]interface{}(nil), r)
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Code) // todo: should be "20" ??
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Key)
	assert.Equal(t, "field 'XXX' not found", err.(*RfcError).ErrorInfo.Message)
	c.Close()
}

//
// STFC Tests
//

func TestFunctionDescription(t *testing.T) {
	fmt.Println("STFC: Get Function Description")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	d, err := c.GetFunctionDescription("STFC_CONNECTION")
	assert.Nil(t, err)
	assert.Equal(t, "ECHOTEXT", d.Parameters[0].Name)
	assert.Equal(t, "RESPTEXT", d.Parameters[1].Name)
	assert.Equal(t, "REQUTEXT", d.Parameters[2].Name)
	c.Close()
}

func TestTableRowAsStructure(t *testing.T) {
	fmt.Println("STFC: Table rows as structure")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	type importedStruct struct {
		RFCFLOAT float64
		RFCCHAR1 string
		RFCCHAR2 string
		RFCCHAR4 string
		RFCINT1  int
		RFCINT2  int
		RFCINT4  int
		RFCHEX3  []byte
		RFCTIME  time.Time
		RFCDATE  time.Time
		RFCDATA1 string
		RFCDATA2 string
	}
	type parameter struct {
		IMPORTSTRUCT importedStruct
		RFCTABLE     []importedStruct
	}
	importStruct := importedStruct{4.23456789, "A", "BC", "DEFG", 1, 2, 345, []byte{0, 11, 12}, time.Now(), time.Now(), "HELLÖ SÄP", "DATA222"}
	params := parameter{importStruct, []importedStruct{importStruct}}
	r, err := c.Call("STFC_STRUCTURE", params)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	assert.Nil(t, r["IMPORTSTUCT"])

	assert.NotNil(t, r["ECHOSTRUCT"])
	echoStruct := r["ECHOSTRUCT"].(map[string]interface{})
	assert.Equal(t, importStruct.RFCFLOAT, echoStruct["RFCFLOAT"])
	assert.Equal(t, importStruct.RFCCHAR1, echoStruct["RFCCHAR1"])
	assert.Equal(t, importStruct.RFCCHAR2, echoStruct["RFCCHAR2"])
	assert.Equal(t, importStruct.RFCCHAR4, echoStruct["RFCCHAR4"])
	assert.Equal(t, importStruct.RFCINT1, echoStruct["RFCINT1"])
	assert.Equal(t, importStruct.RFCINT2, echoStruct["RFCINT2"])
	assert.Equal(t, importStruct.RFCINT4, echoStruct["RFCINT4"])
	assert.Equal(t, importStruct.RFCHEX3, echoStruct["RFCHEX3"])
	assert.Equal(t, importStruct.RFCTIME.Format("150405"), echoStruct["RFCTIME"].(time.Time).Format("150405"))
	assert.Equal(t, importStruct.RFCDATE.Format("20060102"), echoStruct["RFCDATE"].(time.Time).Format("20060102"))
	assert.Equal(t, importStruct.RFCDATA1, echoStruct["RFCDATA1"])
	assert.Equal(t, importStruct.RFCDATA2, echoStruct["RFCDATA2"])

	assert.NotNil(t, r["RFCTABLE"])
	echoTableLine := r["RFCTABLE"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, importStruct.RFCFLOAT, echoTableLine["RFCFLOAT"])
	assert.Equal(t, importStruct.RFCCHAR1, echoTableLine["RFCCHAR1"])
	assert.Equal(t, importStruct.RFCCHAR2, echoTableLine["RFCCHAR2"])
	assert.Equal(t, importStruct.RFCCHAR4, echoTableLine["RFCCHAR4"])
	assert.Equal(t, importStruct.RFCINT1, echoTableLine["RFCINT1"])
	assert.Equal(t, importStruct.RFCINT2, echoTableLine["RFCINT2"])
	assert.Equal(t, importStruct.RFCINT4, echoTableLine["RFCINT4"])
	assert.Equal(t, importStruct.RFCHEX3, echoTableLine["RFCHEX3"])
	assert.Equal(t, importStruct.RFCTIME.Format("150405"), echoTableLine["RFCTIME"].(time.Time).Format("150405"))
	assert.Equal(t, importStruct.RFCDATE.Format("20060102"), echoTableLine["RFCDATE"].(time.Time).Format("20060102"))
	assert.Equal(t, importStruct.RFCDATA1, echoTableLine["RFCDATA1"])
	assert.Equal(t, importStruct.RFCDATA2, echoTableLine["RFCDATA2"])
	c.Close()
}

func TestTableRowAsMap(t *testing.T) {
	fmt.Println("STFC: Table rows as saps")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)

	params := map[string]interface{}{
		"IMPORTSTRUCT": map[string]interface{}{
			"RFCFLOAT": 1.23456789,
			"RFCCHAR1": "A",
			"RFCCHAR2": "BC",
			"RFCCHAR4": "ÄBC",
			"RFCINT1":  0xfe,
			"RFCINT2":  0x7ffe,
			"RFCINT4":  999999999,
			"RFCHEX3":  []byte{255, 254, 253},
			"RFCTIME":  time.Now(),
			"RFCDATE":  time.Now(),
			"RFCDATA1": "HELLÖ SÄP",
			"RFCDATA2": "DATA222",
		},
	}
	r, _ := c.Call("STFC_STRUCTURE", params)

	assert.NotNil(t, r["ECHOSTRUCT"])
	importStruct := params["IMPORTSTRUCT"].(map[string]interface{})
	echoStruct := r["ECHOSTRUCT"].(map[string]interface{})
	assert.Equal(t, importStruct["RFCFLOAT"], echoStruct["RFCFLOAT"])
	assert.Equal(t, importStruct["RFCCHAR1"], echoStruct["RFCCHAR1"])
	assert.Equal(t, importStruct["RFCCHAR2"], echoStruct["RFCCHAR2"])
	assert.Equal(t, importStruct["RFCCHAR4"], echoStruct["RFCCHAR4"])
	assert.Equal(t, importStruct["RFCINT1"], echoStruct["RFCINT1"])
	assert.Equal(t, importStruct["RFCINT2"], echoStruct["RFCINT2"])
	assert.Equal(t, importStruct["RFCINT4"], echoStruct["RFCINT4"])
	assert.Equal(t, importStruct["RFCHEX3"], echoStruct["RFCHEX3"])
	assert.Equal(t, importStruct["RFCTIME"].(time.Time).Format("150405"), echoStruct["RFCTIME"].(time.Time).Format("150405"))
	assert.Equal(t, importStruct["RFCDATE"].(time.Time).Format("20060102"), echoStruct["RFCDATE"].(time.Time).Format("20060102"))
	assert.Equal(t, importStruct["RFCDATA1"], echoStruct["RFCDATA1"])
	assert.Equal(t, importStruct["RFCDATA2"], echoStruct["RFCDATA2"])
	c.Close()
}

func TestTableRowAsVariable(t *testing.T) {
	fmt.Println("STFC: Table rows as single variables")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)

	// array of byte sequences
	certTable := [][]byte{
		[]byte("ABC"),
		[]byte("DEF"),
	}
	params := map[string]interface{}{
		"IT_CERTLIST": certTable,
	}
	r, err := c.Call("SSFR_PSE_CREATE", params)
	assert.Nil(t, err)
	bapiret := r["ET_BAPIRET2"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, bapiret["ID"], "1S")
	assert.Equal(t, bapiret["MESSAGE"], "Creating PSE failed (INITIAL)")

	// array of maps, works as well, as a workaround
	certTableMap := []map[string]interface{}{
		map[string]interface{}{
			"": []byte("ABC"),
		},
		map[string]interface{}{
			"": []byte("DEF"),
		},
	}
	params = map[string]interface{}{
		"IT_CERTLIST": certTableMap,
	}
	r, err = c.Call("SSFR_PSE_CREATE", params)
	assert.Nil(t, err)
	// same error message
	bapiret = r["ET_BAPIRET2"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, bapiret["ID"], "1S")
	assert.Equal(t, bapiret["MESSAGE"], "Creating PSE failed (INITIAL)")
	c.Close()
}

func TestConfigParameter(t *testing.T) {
	fmt.Println("STFC: Connection options: rstrip, returnImportParams")
	//rstrip = false
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	c.RStrip(false)
	r, _ := c.Call("STFC_CONNECTION", map[string]interface{}{"REQUTEXT": "HELLÖ SÄP"})
	assert.Equal(t, 257, len(reflect.ValueOf(r["ECHOTEXT"]).String()))
	assert.Equal(t, "HELLÖ SÄP", strings.TrimSpace(reflect.ValueOf(r["ECHOTEXT"]).String()))

	//returnImportParams = true
	c, _ = ConnectionFromParams(abapSystem())
	c.ReturnImportParams(true)
	r, _ = c.Call("STFC_CONNECTION", map[string]interface{}{"REQUTEXT": "HELLÖ SÄP"})
	assert.Equal(t, "HELLÖ SÄP", r["REQUTEXT"])
	c.Close()
}

func TestInvalidParameterFunctionCall(t *testing.T) {
	fmt.Println("STFC: Invalid RFM parameter")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)
	r, err := c.Call("STFC_CONNECTION", map[string]interface{}{"XXX": "wrongParameter"})
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Equal(t, "Could not get the parameter description for \"XXX\"", err.(*RfcError).Description)
	assert.Equal(t, "field 'XXX' not found", err.(*RfcError).ErrorInfo.Message)
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Code)
	assert.Equal(t, "RFC_INVALID_PARAMETER", err.(*RfcError).ErrorInfo.Key)
	c.Close()
}

//
// Error test
//

func TestErrorFunctionCall(t *testing.T) {
	fmt.Println("Error: ABAP message")
	c, err := ConnectionFromParams(abapSystem())
	assert.Nil(t, err)

	r, err := c.Call("RFC_RAISE_ERROR", map[string]interface{}{"MESSAGETYPE": "A"})
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Equal(t, "Could not invoke function \"RFC_RAISE_ERROR\"", err.(*RfcError).Description)
	assert.Equal(t, "Function not supported", err.(*RfcError).ErrorInfo.Message)
	assert.Equal(t, "RFC_ABAP_MESSAGE", err.(*RfcError).ErrorInfo.Code)
	assert.Equal(t, "Function not supported", err.(*RfcError).ErrorInfo.Key)
	assert.Equal(t, "SR", err.(*RfcError).ErrorInfo.AbapMsgClass)
	assert.Equal(t, "A", err.(*RfcError).ErrorInfo.AbapMsgType)
	assert.Equal(t, "006", err.(*RfcError).ErrorInfo.AbapMsgNumber)
	assert.Equal(t, "STRING", err.(*RfcError).ErrorInfo.AbapMsgV1)
	c.Close()
}

func abapSystem() ConnectionParameters {
	return ConnectionParameters{
		"user":   "demo",
		"passwd": "welcome",
		"ashost": "10.68.110.51",
		"sysnr":  "00",
		"client": "620",
		"lang":   "EN",
	}
}
