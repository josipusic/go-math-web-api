package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

var cache = make(map[string]*result)

var paths = map[string]string{
    "add": "/add",
    "sub": "/subtract",
    "mul": "/multiply",
    "div": "/divide",
}

type result struct {
	Action     string `json:"action"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Answer     string `json:"answer"`
	CacheUsed  bool   `json:"cacheUsed"`
}

func main() {
    for key := range paths {
	    http.HandleFunc(paths[key], handler)
    }
	http.ListenAndServe(":8080", nil)
}

func writeErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseAndValidateParams(r *http.Request) (int, int, error) {
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")
	path := r.URL.Path

	var errs []error

	marginErrStr := "valid between " + strconv.Itoa(math.MinInt) + " and " + strconv.Itoa(math.MaxInt)

	if yStr == "0" {
		errs = append(errs, errors.New("division by zero. y " + marginErrStr))
	}

	x, x_err := strconv.Atoi(xStr)
	if x_err != nil {
        if xStr == "" {
            errs = append(errs, errors.New("supply value for x. x is integer " + marginErrStr))
        } else {
            errs = append(errs, errors.New("invalid value for x: " + xStr + ". x is integer " + marginErrStr))
        }
	}

	y, y_err := strconv.Atoi(yStr)
	if y_err != nil {
        if yStr == "" {
            errs = append(errs, errors.New("supply value for y. y is integer " + marginErrStr))
        } else {
            errs = append(errs, errors.New("invalid value for y: " + yStr + ". y is integer " + marginErrStr))
        }
	}

    // guard from overflow result
	if y_err == nil && x_err == nil {
        calculation_overflow_err := errors.New("calculation overflow." + " Result " + marginErrStr)
        if path == paths["add"] {
            if  (x > 0 && y > 0 && math.MaxInt - x < y) || (x < 0 && y < 0 && math.MinInt - x > y) {
                errs = append(errs, calculation_overflow_err)
            }
        } else if path == paths["sub"] {
            if  (x > 0 && y < 0 && math.MaxInt - x > y) || (x < 0 && y > 0 && math.MinInt - x < y) {
                errs = append(errs, calculation_overflow_err)
            }
        } else if path == paths["mul"] {
            if (math.MaxInt / x < y) || (math.MaxInt / y < x) {
                errs = append(errs, calculation_overflow_err)
            }
        }
    }

	var newErr error
	for i, err := range errs {
		if i == 0 {
			newErr = err
			continue
		}
		newErr = fmt.Errorf("%v; %v", newErr, err)
	}
	if newErr != nil {
		return 0, 0, newErr
	}

	return x, y, nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	x, y, err := parseAndValidateParams(r)
	if err != nil {
		writeErrorResponse(w, err.Error())
		return
	}

	var res result
	var is_cached bool
    cache_key := getCacheKey(r.URL.Path, x, y)
	if res, is_cached = checkCache(cache_key); !is_cached {
		res = calculate(r.URL.Path, x, y)
		cache[cache_key] = &res
	}
	res.CacheUsed = is_cached
	json.NewEncoder(w).Encode(res)
}

func getCacheKey(path string, x int, y int) string {
    return fmt.Sprintf("%s_%s_%s", path, strconv.Itoa(x), strconv.Itoa(y))
}

func checkCache(key string) (result, bool) {
	res, is_cached := cache[key]
	if is_cached {
		return *res, true
	}
	return result{}, false
}

func expireCache(key string) {
    time.Sleep(time.Minute)
    delete(cache, key)
}

func lStrip(path string) string {
    return fmt.Sprint(path[1:])
}

func calculate(path string, x int, y int) result {
	var res result
	res.X = x
	res.Y = y

	switch path {
    case paths["add"]:
		res.Action = lStrip(paths["add"])
		res.Answer = strconv.Itoa(x + y)
	case paths["sub"]:
		res.Action = lStrip(paths["sub"])
		res.Answer = strconv.Itoa(x - y)
	case paths["mul"]:
		res.Action = lStrip(paths["mul"])
		res.Answer = strconv.Itoa(x * y)
	case paths["div"]:
		res.Action = lStrip(paths["div"])
		res.Answer = strconv.FormatFloat(float64(x) /  float64(y), 'E', -1, 32)
	}

    go expireCache(getCacheKey(path, x, y))

	return res
}
