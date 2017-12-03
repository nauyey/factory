package factory_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/nauyey/factory"
	"github.com/nauyey/factory/def"
)

type testUser struct {
	ID        int64
	Name      string
	NickName  string
	Age       int32
	Country   string
	BirthTime time.Time
	Now       time.Time
	Blogs     []*testBlog
}

type testBlog struct {
	ID       int64
	Title    string
	Content  string
	AuthorID int64
	Author   *testUser
}

type testComment struct {
	ID     int64
	Text   string
	BlogID int64
	UserID int64
	Blog   *testBlog
	User   *testUser
}

type relation struct {
	Author *testUser
}

type testCommentary struct {
	ID       int64
	Title    string
	Content  string
	AuthorID int64
	R        *relation
	Comment  *testComment
}

func TestBuild(t *testing.T) {
	// define factory
	var birthTime, _ = time.Parse("2006-01-02T15:04:05.000Z", "2000-11-19T00:00:00.000Z")
	var now, _ = time.Parse("2006-01-02T15:04:05.000Z", "2017-11-19T00:00:00.000Z")
	userFactory := def.NewFactory(testUser{}, "",
		def.Field("Name", "test name"),
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
		def.Field("Now", now),
		def.Field("BirthTime", birthTime),
		def.DynamicField("Age", func(model interface{}) (interface{}, error) {
			user, ok := model.(*testUser)
			if !ok {
				return nil, errors.New("invalid type of model in DynamicFieldValue function")
			}
			return int32(user.Now.Sub(user.BirthTime).Hours() / (24 * 365)), nil
		}),
		def.Trait("Chinese",
			def.Field("Name", "小明"),
			def.Field("Country", "China"),
		),
		def.Trait("teenager",
			def.Field("Name", "少年小明"),
			def.Field("NickName", "Young Ming"),
			def.Field("Age", int32(16)),
		),
		def.Trait("a year latter",
			def.SequenceField("Age", 1, func(n int64) (interface{}, error) {
				return n, nil
			}),
		),
		def.AfterBuild(func(model interface{}) error {
			fmt.Println("AfterBuild...")
			fmt.Println(model)
			return nil
		}),
	)

	// Test default factory
	user := &testUser{}
	err := Build(userFactory).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkUser(t, "Test default factory",
		&testUser{
			ID:       1,
			Name:     "test name",
			NickName: "",
			Age:      17,
			Country:  "",
		},
		user,
	)

	// Test Build with Field
	user = &testUser{}
	err = Build(userFactory,
		WithField("Name", "Little Baby"),
		WithField("Age", int32(3)),
	).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkUser(t, "Test Build with Field",
		&testUser{
			ID:       2,
			Name:     "Little Baby",
			NickName: "",
			Age:      3,
			Country:  "",
		},
		user,
	)

	// Test Build with Trait
	user = &testUser{}
	err = Build(userFactory, WithTraits("Chinese")).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkUser(t, "Test Build with Trait",
		&testUser{
			ID:       3,
			Name:     "小明",
			NickName: "",
			Age:      17,
			Country:  "China",
		},
		user,
	)

	// Test Build with multi Traits
	user = &testUser{}
	err = Build(userFactory, WithTraits("Chinese", "teenager")).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkUser(t, "Test Build with multi Traits",
		&testUser{
			ID:       4,
			Name:     "少年小明",
			NickName: "Young Ming",
			Age:      16,
			Country:  "China",
		},
		user,
	)

	// Test Build with multi Traits and Field
	user = &testUser{}
	err = Build(userFactory,
		WithTraits("Chinese", "teenager"),
		WithField("Name", "中本聪明"),
	).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkUser(t, "Test Build with multi Traits",
		&testUser{
			ID:       5,
			Name:     "中本聪明",
			NickName: "Young Ming",
			Age:      16,
			Country:  "China",
		},
		user,
	)
}

func TestBuildWithAssociation(t *testing.T) {
	// define user factory
	userFactory := def.NewFactory(testUser{}, "",
		def.Field("Name", "test name"),
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
	)
	// define blog factory
	blogFactory := def.NewFactory(testBlog{}, "",
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
		def.DynamicField("Title", func(blog interface{}) (interface{}, error) {
			blogInstance, ok := blog.(*testBlog)
			if !ok {
				return nil, fmt.Errorf("set field Title failed")
			}
			return fmt.Sprintf("Blog Title %d", blogInstance.ID), nil
		}),
		def.Association("Author", "AuthorID", "ID", userFactory,
			def.Field("Name", "blog author name"),
		),
	)

	// Test Build with association
	blog := &testBlog{}
	err := Build(blogFactory, WithField("Content", "Blog content")).To(blog)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	checkBlog(t, "Test Build with association",
		&testBlog{
			ID:       1,
			Title:    "Blog Title 1",
			Content:  "Blog content",
			AuthorID: 1,
			Author: &testUser{
				ID:       1,
				Name:     "blog author name",
				NickName: "",
				Age:      0,
				Country:  "",
			},
		},
		blog,
	)
}

func TestBuildOneToManyAssociation(t *testing.T) {
	// define blog factory
	blogFactory := def.NewFactory(testBlog{}, "",
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
		def.DynamicField("Title", func(blog interface{}) (interface{}, error) {
			blogInstance, ok := blog.(*testBlog)
			if !ok {
				return nil, fmt.Errorf("set field Title failed")
			}
			return fmt.Sprintf("Blog Title %d", blogInstance.ID), nil
		}),
	)
	// define user factory
	userFactory := def.NewFactory(testUser{}, "",
		def.Field("Name", "test one-to-many name"),
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
		def.AfterBuild(func(user interface{}) error {
			author, _ := user.(*testUser)

			author.Blogs = []*testBlog{}
			return BuildSlice(blogFactory, 10,
				WithField("AuthorID", author.ID),
				WithField("Author", author),
			).To(&author.Blogs)
		}),
	)

	// Test Build one-to-many association
	user := &testUser{}
	err := Build(userFactory).To(user)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	if len(user.Blogs) != 10 {
		t.Fatalf("Build one-to-many association failed with len(Blogs)=%d, want len(Blogs)=10", len(user.Blogs))
	}
	for i, blog := range user.Blogs {
		checkBlog(t, "Test Build one-to-many association",
			&testBlog{
				ID:       int64(i) + 1,
				Title:    fmt.Sprintf("Blog Title %d", i+1),
				AuthorID: user.ID,
				Author: &testUser{
					ID:   user.ID,
					Name: "test one-to-many name",
				},
			},
			blog,
		)
	}
}

func TestBuildWithChainedField(t *testing.T) {
	// define user factory
	userFactory := def.NewFactory(testUser{}, "",
		def.Field("Name", "test name"),
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
	)
	// define commentary factory
	commentaryFactory := def.NewFactory(testCommentary{}, "",
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
		def.DynamicField("Title", func(commentaryIfac interface{}) (interface{}, error) {
			commentary, ok := commentaryIfac.(*testCommentary)
			if !ok {
				return nil, fmt.Errorf("set field Title failed")
			}
			return fmt.Sprintf("Blog Title %d", commentary.ID), nil
		}),
		def.Association("R.Author", "AuthorID", "ID", userFactory,
			def.Field("Name", "commentary author name"),
		),
		def.Trait("with comment text",
			def.Field("Comment.Text", "chain comment text"),
		),
	)

	// Test Build with chained field
	commentary := &testCommentary{}
	err := Build(commentaryFactory, WithTraits("with comment text")).To(commentary)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	if commentary.Comment.Text != "chain comment text" {
		t.Fatalf("Build failed with chained field with Comment.Text=%s, want Comment.Text=\"chain comment text\"", commentary.Comment.Text)
	}

	// Test Build Associations with chained field
	commentary = &testCommentary{}
	err = Build(commentaryFactory).To(commentary)
	if err != nil {
		t.Fatalf("Build failed with error: %v", err)
	}
	if commentary.R == nil {
		t.Fatalf("Build failed with chained field with R=nil")
	}
	if commentary.R.Author == nil {
		t.Fatalf("Build failed with chained field with R.Author=nil")
	}
	if commentary.R.Author.Name != "commentary author name" {
		t.Errorf("Build failed with chained field with R.Author.Name=%s, want R.Author.Name = \"commentary author name\"", commentary.R.Author.Name)
	}
}

func TestBuildSlice(t *testing.T) {
	// define user factory
	userFactory := def.NewFactory(testUser{}, "",
		def.Field("Name", "test name"),
		def.SequenceField("ID", 1, func(n int64) (interface{}, error) {
			return n, nil
		}),
	)

	// test build []*Type slice
	users := []*testUser{}
	err := BuildSlice(userFactory, 3, WithField("Name", "test build slice name")).To(&users)
	if err != nil {
		t.Fatalf("BuildSlice failed with err=%v", err)
	}
	if len(users) != 3 {
		t.Fatalf("BuildSlice failed with len(users)=%d, want len(users)=3", len(users))
	}
	for i, user := range users {
		checkUser(t, "Test BuildSlice",
			&testUser{
				ID:   int64(i) + 1,
				Name: "test build slice name",
			},
			user,
		)
	}

	// test build []Type slice
	users2 := []testUser{}
	err = BuildSlice(userFactory, 3, WithField("Name", "test build slice name")).To(&users2)
	if err != nil {
		t.Fatalf("BuildSlice failed with err=%v", err)
	}
	if len(users2) != 3 {
		t.Fatalf("BuildSlice failed with len(users2)=%d, want len(users2)=3", len(users2))
	}
	for i, user := range users2 {
		checkUser(t, "Test BuildSlice",
			&testUser{
				ID:   int64(i) + 4,
				Name: "test build slice name",
			},
			&user,
		)
	}

	// test build with count=0
	users = []*testUser{}
	err = BuildSlice(userFactory, 0, WithField("Name", "test build slice name")).To(&users)
	if err != nil {
		t.Fatalf("BuildSlice failed with err=%v", err)
	}
	if len(users) != 0 {
		t.Fatalf("BuildSlice failed with len(users)=%d, want len(users)=0", len(users))
	}
}

func checkUser(t *testing.T, name string, expect *testUser, got *testUser) {
	if got.ID != expect.ID {
		t.Errorf("Case %s: failed with ID=%d, want ID=%d", name, got.ID, expect.ID)
	}
	if got.Name != expect.Name {
		t.Errorf("Case %s: failed with Name=%s, want Name=%s", name, got.Name, expect.Name)
	}
	if got.NickName != expect.NickName {
		t.Errorf("Case %s: failed with NickName=%s, want NickName=%s", name, got.NickName, expect.NickName)
	}
	if got.Age != expect.Age {
		t.Errorf("Case %s: failed with Age=%d, want Age=%d", name, got.Age, expect.Age)
	}
	if got.Country != expect.Country {
		t.Errorf("Case %s: failed with Country=%s, want Country=%s", name, got.Country, expect.Country)
	}
}

func checkBlog(t *testing.T, name string, expect *testBlog, got *testBlog) {
	if got.ID != expect.ID {
		t.Errorf("Case %s: failed with ID=%d, want ID=%d", name, got.ID, expect.ID)
	}
	if got.Title != expect.Title {
		t.Errorf("Case %s: failed with Title=%s, want Title=%s", name, got.Title, expect.Title)
	}
	if got.Content != expect.Content {
		t.Errorf("Case %s: failed with Content=%s, want Content=%s", name, got.Content, expect.Content)
	}
	if got.AuthorID != expect.AuthorID {
		t.Errorf("Case %s: failed with AuthorID=%d, want AuthorID=%d", name, got.AuthorID, expect.AuthorID)
	}
	checkUser(t, name, expect.Author, got.Author)
}
