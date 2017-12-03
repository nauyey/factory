Getting Started
===============

Table of Contents
-----------------

* [Table of Contents](#table-of-contents)
* [Defining factories](#defining-factories)
* [Using factories](#using-factories)
* [Chained Field](#chained-field)
* [Dynamic Fields](#dynamic-fields)
* [Dependent Fields](#dependent-fields)
* [Associations](#associations)
* [Sequence Fields](#sequence-fields)
* [Trait](#trait)
* [Callbacks](#callbacks)
* [Building or Creating Multiple Records](#building-or-creating-multiple-records)

Defining factories
------------------

Each factory is a def.Factory instance which is related to a specific golang struct and has a set of fields. For factory to create saved instance, the database table name is also needed:

```golang
import "github.com/nauyey/factory/def"

type User struct {
    ID        int64
    Name      string
    Gender    string
    Age       int
    BirthTime time.Time
    Country   string
	Email     string
}

// This will define a factory for User struct
userFactory := def.NewFactory(User{}, "",
	def.Field("Name", "test name"),
	def.SequenceField("ID", 0, func(n int64) interface{} {
		return n
	}),
	def.Trait("Chinese",
		def.Field("Country", "china"),
	),
	def.AfterBuild(func(user interface{}) error {
		// do something
	}),
)

type UserForSave struct {
    ID        int64     `factory:"id,primary"`
    Name      string    `factory:"name"`
    Gender    string    `factory:"gender"`
    Age       int       `factory:"age"`
    BirthTime time.Time `factory:"birth_time`
    Country   string    `factory:"country"`
    Email     string    `factory:"email"`
}

// This will define a factory for UserForSave struct with database table
userForSaveFactory := def.NewFactory(UserForSave{}, "model_table",
	def.Field("Name", "test name"),
	def.SequenceField("ID", 0, func(n int64) interface{} {
		return n
	}),
	def.Trait("Chinese",
		def.Field("Country", "china"),
	),
	def.BeforeCreate(func(user interface{}) error {
		// do something
	}),
	def.AfterCreate(func(user interface{}) error {
		// do something
	}),
)
```

For factory to create saved instance, the struct fields will be mapped to database table fields by tags declared in the origianl struct. Tag name is `factory`. And the mapping rules are as following:

1. If a struct field, like ID, has tag `factory:"id"`, then the field will be map to be the field "id" in database table.
2. If a struct field, like ID, has tag `factory:"id,primary"`, then the field will be map to table field "id", and factory will treat it as the primary key of the table.
3. If a struct field, like NickName, has tag `factory:""`, `factory:","`, `factory:",primary"` or `factory:",anything else"`, then the field will be map to the table field named "nick_name". In this situation, factory just use the snake case of the original struct field name as table field name.


It is highly recommended that you have one factory for each struct that provides the simplest set of fields necessary to create an instance of that struct.

For different kinds of scenarios, you can define different traits for them.


Using factories
---------------

factory supports several different build strategies: Build, Create, Delete:

```golang
import . "github.com/nauyey/factory"

// Returns a user instance that's not saved
user := &User{}
err := Build(userFactory).To(user)

// Returns a saved User instance
user := &User{}
err := Create(userFactory).To(user)

// Deletes a saved User instance from database
err := Delete(userFactory, user)
```

No matter which strategy is used, it's possible to override the defined fields by passing  `factoryOption` type of parameters. Currently, factory supports `WithTraits`, `WithField`:

```golang
import . "github.com/nauyey/factory"

// Build a User instance and override the name field
user := &User{}
err := Build(userFactory, WithField("Name", "Tony")).To(user)
// user.Name => "Tony"

// Build a User instance with traits
user := &User{}
err := Build(userFactory, 
    WithTraits("Chinese"), 
    WithField("Name", "XiaoMing"),
).To(user)
// user.Name => "XiaoMing"
// user.Country => "China"
```

Before using Create, Delete, a `*sql.DB` instance should be seted to factory:

```golang
import "github.com/nauyey/factory"

var db *sql.DB

// init a *sql.DB instance to db

factory.SetDB(db)
```


Chained Field
-------------

Chained field supplies a way to set nested struct field values:

```golang
import (
    "time"

    "github.com/nauyey/factory/def"
)

type Blog struct {
	ID       int64  `factory:"id,primary"`
	Title    string `factory:"title"`
	Content  string `factory:"content"`
	AuthorID int64  `factory:"author_id"`
	Author   *User
}

blogFactory := def.NewFactory(Blog{}, "",
	def.Field("Title", "blog title"),
	def.Field("Author.Name", "blog author name"),
)
blog := &Blog{}
err := Build(blogFactory).To(blog)
// blog.Title => "blog title"
// blog.Author.Name => "blog author name"
``` 

Dynamic Fields
--------------

Most factory fields can be added using static values that are evaluated when the factory is defined, but some fields (such as associations and other fields that must be dynamically generated) will need values assigned each time an instance is generated. These "dynamic" fields can be added by passing a `DynamicFieldValue` type generator function to `DynamicField` instead of a parameter:

```golang
import (
    "time"

    "github.com/nauyey/factory/def"
)

userFactory := def.NewFactory(User{}, "model_table",
	def.Field("Name", "test name"),
	def.DynamicField("Age", func(model interface{}) interface{} {
        now := time.Now()
        birthTime, _ := time.Parse("2006-01-02T15:04:05.000Z", "2017-11-19T00:00:00.000Z")
		return birthTime.Sub(now).Years()
	}),
)
```

Dependent Fields
----------------

Fields can be based on the values of other fields using the evaluator that is yielded to dynamic field value generator function:

```golang
import (
    "time"

    "github.com/nauyey/factory/def"
)

userFactory := def.NewFactory(User{}, "model_table",
	def.Field("Name", "test name"),
	def.DynamicField("Age", func(model interface{}) (interface{}, error) {
        user, ok := model.(*User)
        if !ok {
            return nil, errors.NewFactory("invalid type of model in DynamicFieldValue function")
        }
        now := time.Now()
		return user.BirthTime.Sub(now).Years()
	}),
)
```

Associations
------------

It's possible to set up associations within factories. Use `def.Association` to define an association of a factory by specify a different def. And you can override fields definitions of this association factory.

```golang
import "github.com/nauyey/factory/def"

type Blog struct {
	ID       int64  `factory:"id,primary"`
	Title    string `factory:"title"`
	Content  string `factory:"content"`
	AuthorID int64  `factory:"author_id"`
	Author   *User
}

userFactory := def.NewFactory(User{}, "user_table",
	def.Field("Name", "test name"),
)

blogFactory := def.NewFactory(Blog{}, "blog_table",
    // define an association
	def.Association("Author", "AuthorID", "ID", userFactory,
		def.Field("Name", "blog author name"), // override field
	),
)
```

In factory, here isn't a direct way to define one-to-many relationships. But you can define a one-to-many relationships in `def.AfterBuild` and `def.AfterCreate`:

```golang
import (
	. "github.com/nauyey/factory"
	"github.com/nauyey/factory/def"
)

type User struct {
	ID    int64  `factory:"id,primary"`
	Name  string `factory:"name"`
	Blogs []*Blog
}

type Blog struct {
	ID       int64  `factory:"id,primary"`
	Title    string `factory:"title"`
	Content  string `factory:"content"`
	AuthorID int64  `factory:"author_id"`
	Author   *User
}

blogFactory := def.NewFactory(Blog{}, "blog_table",
    def.SequenceField("ID", 1, func(n int64) interface{} {
		return n
	}),
)

// define unsaved one-to-many associations in AfterBuild
userFactory := def.NewFactory(User{}, "",
	def.Field("Name", "test name"),
	def.Trait("with unsaved blogs",
		def.AfterBuild(func(user interface{}) error {
			author, _ := user.(*User)

			author.Blogs = []*Blog{}
			return BuildSlice(blogFactory, 10,
				WithField("AuthorID", author.ID),
				WithField("Author", author),
			).To(&author.Blogs)
		}),
	),
)

// define saved one-to-many associations in AfterCreate
userForSaveFactory := def.NewFactory(User{}, "user_table",
	def.Field("Name", "test name"),
	def.Trait("with saved blogs",
		def.AfterCreate(func(user interface{}) error {
			author, _ := user.(*User)
			
			author.Blogs = []*Blog{}
			return CreateSlice(blogFactory, 10,
				WithField("AuthorID", author.ID),
				WithField("Author", author),
			).To(&author.Blogs)
		}),
	),
)
```

The behavior of the `def.Association` function varies depending on the build strategy used for the parent object.

```golang
// Builds and saves a User and a Blog
blog := &Blog{}
err := Create(blogModel).To(blog) // blog is saved into database
user := blog.Author // user is saved into database


// Builds a User and a Blog, but saves nothing
blog := &Blog{}
err := Build(blogModel).To(blog) // blog isn't saved
user = blog.Author // user isn't saved
```

Sequence Fields
---------------

Unique values in a specific format (for example, e-mail addresses) can be generated using sequences. Sequence fields are defined by calling `SequenceField` in factory model defination, and values in a sequence are generated by calling `SequenceFieldValue` type of callback function:

```golang
import (
    . "github.com/nauyey/factory"
    "github.com/nauyey/factory/def"
)

// Defines a new sequence field
userFactory := def.NewFactory(User{}, "model_table",
	def.SequenceField("Email", 0, func(n int64) interface{} {
		return fmt.Sprintf("person%d@example.com", n)
	}),
)

user0 := &User{}
err := Build(userFactory).To(user0)
// user0.Email => "person0@example.com"

user1 := &User{}
err := Build(userFactory).To(user1)
// user1.Email => "person1@example.com"
```

You can also set the initial start of the sequence:

```golang
userFactory := def.NewFactory(User{}, "model_table",
	def.SequenceField("Email", 1000, func(n int64) interface{} {
		return fmt.Sprintf("person%d@example.com", n)
	}),
)

user0 := &User{}
err := Build(userFactory).To(user0)
// user0.Email => "person1000@example.com"
```

Trait
-----

Trait allows you to group fields together and then apply them to the factory model.

```golang
import (
    . "github.com/nauyey/factory"
    "github.com/nauyey/factory/def"
)

userFactory := def.NewFactory(User{}, "",
	def.Field("Name", "Taylor"),
	def.Trait("Chinese boy",
        def.Field("Country", "China"),
        def.Field("Gender", "Male"),
    ),
)

user := &User{}
err := Build(userFactory, WithTraits("Chinese boy")).To(user)
// user.Country => "China"
// user.Gender => "Male"
```


Traits that defines the same fields are allowed.
Traits can also be passed in as a slice of strings, by using `WithTraits`, when you construct an instance from factory.The fields that defined in the latest trait gets precedence.

```golang
import (
    . "github.com/nauyey/factory"
    "github.com/nauyey/factory/def"
)

userFactory := def.NewFactory(User{}, "",
	def.Field("Name", "Taylor"),
	def.Trait("Chinese boy",
        def.Field("Country", "China"),
        def.Field("Gender", "Male"),
    ),
    def.Trait("American",
        def.Field("Country", "USA"),
    ),
    def.Trait("girl",
        def.Field("Gender", "Female"),
    ),
)

user := &User{}
err := Build(userFactory, WithTraits("Chinese boy", "American", "girl")).To(user)
// user.Country => "USA"
// user.Gender => "Female"
```

This ability works with `build` and `create`.


Traits can be used with associations easily too:

```golang
import "github.com/nauyey/factory/def"

blogFactory := def.NewFactory(Blog{}, "blog_table",
    // define an association in traits
    def.Trait("with author",
        def.Association("Author", "AuthorID", "ID", userFactory,
            def.Field("Name", "blog author in trait"), // override field
        ),
    ),
	
)

blog := &Blog{}
err := Build(blogFactory, WithTraits("with author")).To(blog)
// blog.Author
// blog.Author.Name => "blog author in trait"
```

Traits cann't be used within other traits.


Callbacks
---------

factory makes available 3 callbacks for injections:

* `def.AfterBuild`   - called after an instance is built   (via `Build`, `Create`)
* `def.BeforeCreate` - called before an instance is saved  (via `Create`)
* `def.AfterCreate`  - called after an instance is saved   (via `Create`)

Examples:

```golang
import "github.com/nauyey/factory/def"

// Define a factory that calls the callback function after it is built
userFactory := def.NewFactory(User{}, "",
	def.AfterBuild(func(user interface{}) error {
		// do something
	}),
)
```

Note that you'll have an instance of the user in the callback function. This can be useful.

You can also define multiple types of callbacks on the same factory:

```golang
import "github.com/nauyey/factory/def"

// Define a factory that calls the callback function after it is built
userFactory := def.NewFactory(User{}, "",
	def.AfterBuild(func(user interface{}) error {
		// do something
    }),
    def.BeforeCreate(func(user interface{}) error {
		// do something
    }),
    def.AfterCreate(func(user interface{}) error {
		// do something
	}),
)
```

Factories can also define any number of the same kind of callback.  These callbacks will be executed in the order they are specified:

```golang
import "github.com/nauyey/factory/def"

// Define a factory that calls the callback function after it is built
userFactory := def.NewFactory(User{}, "",
	def.AfterBuild(func(user interface{}) error {
		// do something
    }),
    def.AfterBuild(func(user interface{}) error {
		// do something
    }),
    def.AfterBuild(func(user interface{}) error {
		// do something
	}),
)
```

Calling `Create` will invoke both `def.AfterBuild` and `def.AfterCreate` callbacks.

Building or Creating Multiple Records
-------------------------------------

Sometimes, you'll want to create or build multiple instances of a factory at once.

```golang
import . "github.com/nauyey/factory"

users := []*User{}

err := BuildSlice(userFactory, 10).To(users)
err := CreateSlice(userFactory, 10).To(users)
```

To set the fields for each of the factories, you can use `WithField` and `WithTraits` as you normally would.:

```golang
import . "github.com/nauyey/factory"

users := []*User{}

err := BuildSlice(userFactory, 10, WithField("Name", "build slice name")).To(users)
```