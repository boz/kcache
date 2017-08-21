package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"golang.org/x/tools/imports"
)

type joinDef struct {
	SrcName string
	SrcPkg  string
	SrcType string
	DstName string
	DstPkg  string
}

func main() {

	if len(os.Args) != 6 {
		fmt.Fprintf(os.Stderr, "USAGE: %v: <src-name> <src-pkg> <src-type> <dst-name> <dst-pkg>\n", os.Args[0])
		os.Exit(1)
	}

	def := joinDef{
		SrcName: os.Args[1],
		SrcPkg:  os.Args[2],
		SrcType: os.Args[3],
		DstName: os.Args[4],
		DstPkg:  os.Args[5],
	}

	buf := processTemplate(&def)
	buf = processImports(buf)

	if _, err := os.Stdout.Write(buf); err != nil {
		panic(err.Error())
	}
}

func processTemplate(def *joinDef) []byte {
	buf := &bytes.Buffer{}
	if err := joinTemplate.Execute(buf, &def); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func processImports(in []byte) []byte {
	out, err := imports.Process("./join/generated.go", in, nil)
	if err != nil {
		panic(err)
	}
	return out
}

var joinTemplate = template.Must(template.New("join").Parse(`/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/{{.SrcPkg}}"
	"github.com/boz/kcache/types/{{.DstPkg}}"
)

func {{.SrcName}}{{.DstName}}sWith(ctx context.Context,
	srcController {{.SrcPkg}}.Controller,
	dstController {{.DstPkg}}.Publisher,
	filterFn func(...{{.SrcType}}) filter.ComparableFilter) ({{.DstPkg}}.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ {{.SrcType}}) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join({{.SrcPkg}},{{.DstPkg}}: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := {{.SrcPkg}}.BuildHandler().
		OnInitialize(func(objs []{{.SrcType}}) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := {{.SrcPkg}}.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join({{.SrcPkg}},{{.DstPkg}}): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}`))
