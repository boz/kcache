package nsname_test

import (
	"testing"

	"github.com/boz/kcache/nsname"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {

	{
		id, err := nsname.Parse("foo/bar")
		require.NoError(t, err)
		require.Equal(t, "foo", id.Namespace)
		require.Equal(t, "bar", id.Name)
		require.Equal(t, "foo/bar", id.String())
	}

	{
		id, err := nsname.Parse("foo/")
		require.NoError(t, err)
		require.Equal(t, "foo", id.Namespace)
		require.Equal(t, "", id.Name)
		require.Equal(t, "foo/", id.String())
	}

	{
		id, err := nsname.Parse("/bar")
		require.NoError(t, err)
		require.Equal(t, "", id.Namespace)
		require.Equal(t, "bar", id.Name)
		require.Equal(t, "/bar", id.String())
	}

	{
		id, err := nsname.Parse("/")
		require.NoError(t, err)
		require.Equal(t, "", id.Namespace)
		require.Equal(t, "", id.Name)
		require.Equal(t, "/", id.String())
	}

	{
		_, err := nsname.Parse("/bar/")
		require.Equal(t, nsname.ErrInvalidID, err)
	}

	{
		_, err := nsname.Parse("")
		require.Equal(t, nsname.ErrInvalidID, err)
	}

}
