package def_test

import (
	"strings"
	"testing"

	"github.com/nauyey/factory/def"
)

type testUser struct {
	ID       int64
	Name     string
	NickName string
	Age      int32
	Country  string
}

type testBlog struct {
	ID       int64
	Title    string
	Content  string
	AuthorID int64
	Author   *testUser
}

func TestDuplicateDefinition(t *testing.T) {
	const duplicateDefinitionErr = "duplicate definition of field"

	// test def.Field
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		def.NewFactory(testUser{}, "",
			def.Field("Name", "test name"),
			def.Field("Name", "test name 2"),
		)
	})()

	// test def.SequenceField
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		def.NewFactory(testUser{}, "",
			def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
				return n, nil
			}),
			def.SequenceField("ID", 10, func(n int64) (interface{}, error) {
				return n, nil
			}),
		)
	})()

	// test def.DynamicField
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		def.NewFactory(testUser{}, "",
			def.DynamicField("Age", func(model interface{}) (interface{}, error) {
				return 16, nil
			}),
			def.DynamicField("Age", func(model interface{}) (interface{}, error) {
				return 16, nil
			}),
		)
	})()

	// test def.Association
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		// define user factory
		userFactory := def.NewFactory(testUser{}, "")

		def.NewFactory(testBlog{}, "",
			def.Association("Author", "AuthorID", "ID", userFactory,
				def.Field("Name", "blog author name"),
			),
			def.Association("Author", "AuthorID", "ID", userFactory,
				def.Field("Name", "blog author name"),
			),
		)
	})()

	// test mixed duplication
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		def.NewFactory(testUser{}, "",
			def.Field("ID", int64(20)),
			def.SequenceField("ID", 10, func(n int64) (interface{}, error) {
				return n, nil
			}),
		)
	})()

	// test duplication in def.Trait
	(func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatalf("def.NewFactory should panic by duplicate field definition")
			}
			if ok := strings.Contains(err.(error).Error(), duplicateDefinitionErr); !ok {
				t.Fatalf("expects err: \"%s\" contains \"%s\"", err.(error).Error(), duplicateDefinitionErr)
			}
		}()

		def.NewFactory(testUser{}, "",
			def.Trait("Chinese",
				def.Field("Name", "小明"),
				def.Field("Name", "test name"),
				def.Field("Country", "China"),
			),
		)
	})()

	// test def.Trait overrides definitions in def.NewFactory will not panic
	def.NewFactory(testUser{}, "",
		def.Field("Name", "test name"),
		def.Trait("Chinese",
			def.Field("Name", "小明"),
			def.Field("Country", "China"),
		),
	)
}
