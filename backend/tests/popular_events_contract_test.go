package tests

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	applicationlogger "friendship/logger"
	"friendship/models/dto"
	servicesevents "friendship/services/events"
)

func TestPopularEventsCleanSourceDoesNotImportInfrastructureLibraries(t *testing.T) {
	t.Helper()

	eventsDir := filepath.Join("..", "services", "events")
	files, err := parser.ParseDir(token.NewFileSet(), eventsDir, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse event service sources: %v", err)
	}

	var cleanFiles []string
	for _, pkg := range files {
		for path, file := range pkg.Files {
			fullFile, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("parse event service source %s: %v", path, err)
			}
			if declaresCleanPopularEventsContract(fullFile) {
				cleanFiles = append(cleanFiles, path)
				assertPopularEventsImportsAreClean(t, path, file)
			}
		}
	}

	if len(cleanFiles) == 0 {
		t.Fatal("no clean popular-events service or port source declarations found")
	}
}

func TestNewPopularEventsServiceAcceptsOnlyCleanPorts(t *testing.T) {
	constructorType := reflect.TypeOf(servicesevents.NewPopularEventsService)
	loggerType := reflect.TypeOf((*applicationlogger.Logger)(nil)).Elem()
	serviceType := reflect.TypeOf((*servicesevents.PopularEventsService)(nil)).Elem()

	if constructorType.NumIn() < 2 {
		t.Fatalf("NewPopularEventsService has %d inputs, want logger and clean popular-events ports", constructorType.NumIn())
	}
	if constructorType.In(0) != loggerType {
		t.Fatalf("NewPopularEventsService input 0 = %v, want %v", constructorType.In(0), loggerType)
	}

	checked := make(map[reflect.Type]bool)
	for index := 0; index < constructorType.NumIn(); index++ {
		assertPopularEventsTypeDoesNotLeakInfrastructure(t, constructorType.In(index), checked)
	}
	for index := 0; index < constructorType.NumOut(); index++ {
		assertPopularEventsTypeDoesNotLeakInfrastructure(t, constructorType.Out(index), checked)
	}

	if constructorType.NumOut() == 0 || !constructorType.Out(0).Implements(serviceType) {
		t.Fatalf("NewPopularEventsService result 0 = %v, want implementation of %v", constructorType.Out(0), serviceType)
	}
}

func TestPopularEventsServiceMethodsPropagateContextByContract(t *testing.T) {
	serviceType := reflect.TypeOf((*servicesevents.PopularEventsService)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	for _, methodName := range []string{"GetPopularEvents", "UpdateCache"} {
		method, exists := serviceType.MethodByName(methodName)
		if !exists {
			t.Fatalf("PopularEventsService is missing %s", methodName)
		}
		if method.Type.NumIn() == 0 || method.Type.In(0) != contextType {
			t.Fatalf("%s input 0 = %v, want context.Context", methodName, firstMethodInput(method.Type))
		}
	}
}

func TestCachedPopularEventsUsesSearchItemsWithSubscriptionState(t *testing.T) {
	cachedType := reflect.TypeOf(dto.CachedPopularEvents{})
	eventsField, exists := cachedType.FieldByName("Events")
	if !exists {
		t.Fatal("CachedPopularEvents is missing Events")
	}

	wantEventsType := reflect.TypeOf([]dto.EventSearchItemDto{})
	if eventsField.Type != wantEventsType {
		t.Fatalf("CachedPopularEvents.Events type = %v, want %v", eventsField.Type, wantEventsType)
	}

	itemType := reflect.TypeOf(dto.EventSearchItemDto{})
	for _, field := range []struct {
		name     string
		typeOf   reflect.Type
		jsonName string
	}{
		{name: "StartTime", typeOf: reflect.TypeOf(time.Time{}), jsonName: "startTime"},
		{name: "AgeLimit", typeOf: reflect.TypeOf(""), jsonName: "ageLimit"},
		{name: "Status", typeOf: reflect.TypeOf(""), jsonName: "status"},
	} {
		actual, exists := itemType.FieldByName(field.name)
		if !exists || actual.Type != field.typeOf || actual.Tag.Get("json") != field.jsonName {
			t.Fatalf("EventSearchItemDto.%s = %#v, want %v with json tag %s", field.name, actual, field.typeOf, field.jsonName)
		}
	}

	groupType := reflect.TypeOf(dto.EventSearchGroupDto{})
	groupImage, exists := groupType.FieldByName("Image")
	if !exists || groupImage.Type.Kind() != reflect.String || groupImage.Tag.Get("json") != "image" {
		t.Fatalf("EventSearchGroupDto.Image = %#v, want string with json tag image", groupImage)
	}

	subscribedField, exists := itemType.FieldByName("Subscribed")
	if !exists || subscribedField.Type.Kind() != reflect.Bool || subscribedField.Tag.Get("json") != "subscribed" {
		t.Fatalf("EventSearchItemDto.Subscribed = %#v, want bool with json tag subscribed", subscribedField)
	}
}

func declaresCleanPopularEventsContract(file *ast.File) bool {
	for _, declaration := range file.Decls {
		switch node := declaration.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == "NewPopularEventsService" {
				return true
			}
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == "popularEventsService" {
					return true
				}
				if strings.HasPrefix(typeSpec.Name.Name, "PopularEvent") {
					if _, isInterface := typeSpec.Type.(*ast.InterfaceType); isInterface {
						return true
					}
				}
			}
		}
	}
	return false
}

func assertPopularEventsImportsAreClean(t *testing.T, sourcePath string, file *ast.File) {
	t.Helper()

	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import %s in %s: %v", importSpec.Path.Value, sourcePath, err)
		}
		for _, forbidden := range []string{
			"friendship/config",
			"friendship/email",
			"friendship/models/events",
			"friendship/repository",
			"friendship/utils",
			"github.com/redis/",
			"github.com/robfig/cron",
			"gorm.io/",
		} {
			if importPath == forbidden || strings.HasPrefix(importPath, forbidden) {
				t.Fatalf("%s imports infrastructure dependency %q", sourcePath, importPath)
			}
		}
	}
}

func assertPopularEventsTypeDoesNotLeakInfrastructure(t *testing.T, typ reflect.Type, checked map[reflect.Type]bool) {
	t.Helper()

	if typ == nil || checked[typ] {
		return
	}
	checked[typ] = true

	typeName := typ.String()
	packagePath := typ.PkgPath()
	for _, forbidden := range []string{
		"friendship/config",
		"friendship/email",
		"friendship/models/events",
		"friendship/repository",
		"github.com/redis/",
		"github.com/robfig/cron",
		"gorm.io/",
	} {
		if packagePath == forbidden ||
			strings.HasPrefix(packagePath, forbidden) ||
			strings.Contains(typeName, forbidden) {
			t.Fatalf("popular-events contract leaks infrastructure type %v from %s", typ, packagePath)
		}
	}

	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Elem(), checked)
	case reflect.Map:
		assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Key(), checked)
		assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Elem(), checked)
	case reflect.Func:
		for index := 0; index < typ.NumIn(); index++ {
			assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.In(index), checked)
		}
		for index := 0; index < typ.NumOut(); index++ {
			assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Out(index), checked)
		}
	case reflect.Interface:
		for index := 0; index < typ.NumMethod(); index++ {
			assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Method(index).Type, checked)
		}
	case reflect.Struct:
		if packagePath != "friendship/services/events" {
			return
		}
		for index := 0; index < typ.NumField(); index++ {
			assertPopularEventsTypeDoesNotLeakInfrastructure(t, typ.Field(index).Type, checked)
		}
	}
}

func firstMethodInput(methodType reflect.Type) interface{} {
	if methodType.NumIn() == 0 {
		return "<none>"
	}
	return methodType.In(0)
}
