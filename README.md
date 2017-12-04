Factory: Factory for Go Tests
================================

Factory is a fixtures replacement. With its readable APIs, you can define factories and use factories to create saved, unsaved and stubbed instances by build multiple strategies.

Factory's APIs are inspired by [factory_bot](https://github.com/thoughtbot/factory_bot) in Ruby.

See how easily to use factory:
```golang
import (
	. "github.com/nauyey/factory"
	"github.com/nauyey/factory/def"
)

type User struct {
	ID        int64     `factory:"id,primary"`
	Name      string    `factory:"name"`
	Gender    string    `factory:"gender"`
	Email     string    `factory:"email"`
}

// Define a factory for User struct
userFactory := def.NewFactory(User{}, "db_table_users",
	def.SequenceField("ID", 1, func(n int64) interface{} {
		return n
	}),
	def.DynamicField("Name", func(user interface{}) (interface{}, error) {
		return fmt.Sprintf("User Name %d", user.(*User).ID), nil
	}),
	def.Trait("boy",
		def.Field("Gender", "male"),
	),
)

user := &User{}
Build(userFactory).To(user)
// user.ID   -> 1
// user.Name -> "User Name 1"

user2 := &User{}
Create(userFactory, WithTraits("boy")).To(user2) // saved to database
// user2.ID      -> 2
// user2.Name   -> "User Name 2"
// user2.Gender -> "male"
```

Feature Support
---------------

* Fields
* Dynamic Fields
* Dependent Fields
* Sequence Fields
* Multilevel Fields
* Associations
* Traits
* Callbacks
* Multiple Build Strategies

Installation
------------

Simple install the package to your `$GOPATH` with the go tool from shell:

```bash
$ go get -u github.com/nauyey/factory
```

Documentation
-------------

See 💰💰💰[GETTING_STARTED](GETTING_STARTED.md)💰💰💰 for information on defining and using factories.

How to Contribute
-----------------

1. Check for open issues or open a fresh issue to start a discussion around a feature idea or a bug.
2. Fork [the repository](http://github.com/nauyey/factory) on GitHub to start making your changes to the **master** branch (or branch off of it).
3. Write a test which shows that the bug was fixed or that the feature works as expected.
4. Send a pull request and bug the maintainer until it gets merged and published. :) Make sure to add yourself to [AUTHORS](AUTHORS.md).
