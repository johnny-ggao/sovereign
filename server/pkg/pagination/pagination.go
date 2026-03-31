package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Page    int
	PerPage int
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Params) Limit() int {
	return p.PerPage
}

func Parse(c *gin.Context) Params {
	page := queryInt(c, "page", 1)
	perPage := queryInt(c, "per_page", 20)

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	return Params{Page: page, PerPage: perPage}
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	s := c.Query(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
