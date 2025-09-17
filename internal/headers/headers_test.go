package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersRead(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069                           \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 57, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = map[string]string{"host": "localhost:42069"}
	data = []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 25, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n a bunch of other stuff")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Empty(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: multiple values
	headers = NewHeaders()
	data = []byte("Set-Person: lane-loves-go\r\n")
	data1 := []byte("Set-Person: prime-loves-zig\r\n")
	data2 := []byte("Set-Person: tj-loves-ocaml\r\n")

	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, len(data), n)
	assert.False(t, done)

	n, done, err = headers.Parse(data1)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, len(data1), n)
	assert.False(t, done)

	n, done, err = headers.Parse(data2)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, len(data2), n)
	assert.False(t, done)

	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])

}


