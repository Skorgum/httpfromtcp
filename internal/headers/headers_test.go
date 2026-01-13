package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestSingleHeaderWithExtraWhitespace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("                Host:     localhost:42069                       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 66, n)
	assert.False(t, done)
}

func TestValidTwoHeadersWithExistingHeaders(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nUser-Agent: BootdevClient\r\n\r\n")

	n1, done1, err1 := headers.Parse(data)
	require.NoError(t, err1)
	assert.False(t, done1)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n1)

	n2, done2, err2 := headers.Parse(data[n1:])
	require.NoError(t, err2)
	assert.False(t, done2)
	assert.Equal(t, "BootdevClient", headers["user-agent"])
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 27, n2)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 2, n)
}

func TestCapitolLettersInHeaderKeys(t *testing.T) {
	headers := NewHeaders()
	data := []byte("HoST: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestInvalidCharacterInHeaderKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestHeadersWithMultipleValues(t *testing.T) {
	headers := NewHeaders()
	data := []byte(
		"Set-Person: lane-loves-go\r\n" +
			"Set-Person: prime-loves-zig\r\n" +
			"Set-Person: tj-loves-ocaml\r\n" +
			"\r\n",
	)

	offset := 0
	for {
		n, done, err := headers.Parse(data[offset:])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		offset += n
		if done {
			break
		}
	}

	actual := headers["set-person"]
	expected := "lane-loves-go, prime-loves-zig, tj-loves-ocaml"

	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
