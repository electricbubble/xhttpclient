package xhttpclient

import (
	"cmp"
	"errors"
	"net/url"
)

type xQuery struct {
	query url.Values
}

func (q *xQuery) Query(query url.Values) {
	q.query = query
}

func (q *xQuery) QuerySet(key, value string) {
	if q.query == nil {
		q.query = make(url.Values)
	}
	q.query.Set(key, value)
}

func (q *xQuery) QueryAdd(key, value string) {
	if q.query == nil {
		q.query = make(url.Values)
	}
	q.query.Add(key, value)
}

func (q *xQuery) encode(u, patch *url.URL) (s string, err error) {
	uRawQuery, pRawQuery, ql := u.RawQuery, "", len(q.query)
	if patch != nil {
		pRawQuery = patch.RawQuery
	}

	switch {
	case uRawQuery == "" && pRawQuery == "" && ql == 0:
		return "", nil

	case uRawQuery != "" && pRawQuery == "" && ql == 0:
		return uRawQuery, nil

	case uRawQuery == "" && pRawQuery != "" && ql == 0:
		return pRawQuery, nil

	case uRawQuery == "" && pRawQuery == "":
		return q.query.Encode(), nil

	case (uRawQuery != "") != (pRawQuery != ""):
		final, err := url.ParseQuery(cmp.Or(uRawQuery, pRawQuery))
		if err != nil {
			return "", err
		}
		for k, vv := range q.query {
			for _, v := range vv {
				final.Add(k, v)
			}
		}
		return final.Encode(), nil

	default:
		final, err1 := url.ParseQuery(uRawQuery)
		tmp, err2 := url.ParseQuery(pRawQuery)
		if err = errors.Join(err1, err2); err != nil {
			return "", err
		}
		for k, vv := range tmp {
			for _, v := range vv {
				final.Add(k, v)
			}
		}
		for k, vv := range q.query {
			for _, v := range vv {
				final.Add(k, v)
			}
		}
		return final.Encode(), nil
	}
}
