package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModTimeExpression(t *testing.T) {
	_, err := NewModTimeBefore("")
	assert.ErrorContains(t, err, "failed to parse the date/time expression")

	_, err = NewModTimeBefore("1984/01/23 13:42:23")
	assert.ErrorContains(t, err, "failed to parse the date/time expression")

	_, err = NewModTimeAfter("42D")
	assert.ErrorContains(t, err, "date/time search does not allow shorthand suffixes when using 'after' option")

	now := time.Now().Round(time.Second)

	s, err := NewModTimeBefore("1984-01-23")
	require.NoError(t, err)
	assert.Equal(t, time.Date(1984, 1, 23, 0, 0, 0, 0, time.UTC), s.reference)

	s, err = NewModTimeBefore("13:42:23")
	require.NoError(t, err)
	assert.Equal(t, time.Date(now.Year(), now.Month(), now.Day(), 13, 42, 23, 0, time.UTC), s.reference)

	s, err = NewModTimeBefore("1984-01-23 13:42:23")
	require.NoError(t, err)
	assert.Equal(t, time.Date(1984, 1, 23, 13, 42, 23, 0, time.UTC), s.reference)

	s, err = NewModTimeBefore("1984-01-23T13:42:23")
	require.NoError(t, err)
	assert.Equal(t, time.Date(1984, 1, 23, 13, 42, 23, 0, time.UTC), s.reference)

	s, err = NewModTimeAfter("1984-01-23T13:42:23")
	require.NoError(t, err)
	assert.Equal(t, time.Date(1984, 1, 23, 13, 42, 23, 0, time.UTC), s.reference)

	s, err = NewModTimeBefore("142s")
	require.NoError(t, err)
	assert.Equal(t, now.Add(time.Second*-142), s.reference)

	s, err = NewModTimeBefore("142m")
	require.NoError(t, err)
	assert.Equal(t, now.Add(time.Minute*-142), s.reference)

	s, err = NewModTimeBefore("142h")
	require.NoError(t, err)
	assert.Equal(t, now.Add(time.Hour*-142), s.reference)

	s, err = NewModTimeBefore("42D")
	require.NoError(t, err)
	assert.Equal(t, now.AddDate(0, 0, -42), s.reference)

	s, err = NewModTimeBefore("42M")
	require.NoError(t, err)
	assert.Equal(t, now.AddDate(0, -42, 0), s.reference)

	s, err = NewModTimeBefore("42Y")
	require.NoError(t, err)
	assert.Equal(t, now.AddDate(-42, 0, 0), s.reference)
}
