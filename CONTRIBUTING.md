# Contributing

Wow, we really appreciate that you even looked at this section! We are trying to make the world's best atomic building blocks for financial services that accelerate innovation in banking and we need your help!

You only have a fresh set of eyes once! The easiest way to contribute is to give feedback on the documentation that you are reading right now. This can be as simple as sending a message on our [Slack Channel](https://slack.moov.io) with your feedback or updating the markdown in this documentation and issuing a pull request.

Stability is the hallmark of any good software. If you find an edge case that isn't handled please open an GitHub issue with the example data so that we can make our software more robust for everyone. We also welcome pull requests if you want to get your hands dirty.

Have a use case that we don't handle, or handle well? Start the discussion on our [Slack Channel](https://slack.moov.io) or open a GitHub Issue. We want to make the project meet the needs of the community and keeps you using our code.

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) to ensure you agree with the values of this project.

We use GitHub to manage reviews of pull requests.

* If you have a trivial fix or improvement, go ahead and create a pull request with a description of the change.

* If you plan to do something more involved, first propose your ideas in a Github issue. This will avoid unnecessary work and surely give you and us a good deal of inspiration.

* Relevant coding style guidelines are the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](http://peter.bourgon.org/go-in-production/#formatting-and-style).

* When in doubt follow the [Go Proverbs](https://go-proverbs.github.io/)

* Checkout this [Overview of Go Tooling](https://www.alexedwards.net/blog/an-overview-of-go-tooling) by Alex Edwards

## Getting the code

We recommend using additional git remote's for pushing/pulling code. Go cares about where the `ach` project lives relative to `GOPATH`.

First, pull down our source code:

```
$ git clone git@github.com:moov-io/ach.git
```

Then, add your (or another user's) fork.

```
$ cd $GOPATH/src/github.com/moov-io/ach

$ git remote add $user git@github.com:$user/ach.git

$ git fetch $user
```

Now, feel free to branch and push (`git push $user $branch`) to your remote and send us Pull Requests!

## Pull Requests

A good quality PR will have the following characteristics:

* It will be a complete piece of work that adds value in some way.
* It will have a title that reflects the work within, and a summary that helps to understand the context of the change.
* There will be well written commit messages, with well crafted commits that tell the story of the development of this work.
* Ideally it will be small and easy to understand. Single commit PRs are usually easy to submit, review, and merge.
* The code contained within will meet the best practices set by the team wherever possible.
* The code is able to be merged.
* A PR does not end at submission though. A code change is not made until it is merged and used in production.

A good PR should be able to flow through a peer review system easily and quickly.

Our Build pipeline utilizes [Github Actions](https://github.com/moov-io/ach/actions) to enforce many tools that you should add to your editor before issuing a pull request.

## Additional SEC (Standard Entry Class) code batch types.

SEC types in the Batch Header record define the payment type of the following Entry Details and Addenda. The format of the records in the batch is the same between all payment types but NACHA defines different rules for the values that are held in each record field. To add support for an additional SEC type you will need to implement NACHA rules for that type. The vast majority of rules are implemented in ach.batch and then composed into Batch(SEC) for reuse. All Batch(SEC) types must be an ach.Batcher.

1. Create an issue with the NACHA rules and record layout for the batch type.
2. Create a new struct of the batch type. In the following example we will use MTE (Machine Transfer Entry) as our example.
3. The following code would be place in a new file batchMTE.go next to the existing batch types.
4. The code is stub code and the MTE type is not implemented. For concrete examples review the existing batch types in the source.

Create a new struct and compose ach.batch

```go
type BatchMTE struct {
	batch
}
```
Add the ability for the new type to be created.

```go
func NewBatchMTE(bh *BatchHeader) *BatchMTE {
	batch := new(BatchMTE)
	batch.setControl(NewBatchControl)
	batch.SetHeader(bh)
	batch.SetID(bh.ID)
	return batch
}
```

To support the Batcher interface you must add the following functions that are not implemented in `ach.Batch`.

- `Validate() error`
- `Create() error`

Validate is designed to enforce the NACHA rules for the MTE payment type. Validate is run after a batch of this type is read from a file. If you are creating a batch from code, call validate afterwards.

```go
// Validate checks properties of the ACH batch to ensure they match NACHA guidelines.
// This includes computing checksums, totals, and sequence orderings.
//
// Validate will never modify the batch.
func (batch *BatchMTE) Validate() error {
	// basic verification of the batch before we validate specific rules.
	if err := batch.verify(); err != nil {
		return err
	}
	// Add configuration based validation for this type.
	// ... batch.isAddendaCount(1)
	// Add type specific validation.
	// ...
	return nil
}
```
Create takes the Batch Header and Entry details and creates the proper sequence number and batch control. If additional logic specific to the SEC type is required for building a batch file it should be added here.

```go
// Create will tabulate and assemble an ACH batch into a valid state. This includes
// setting any posting dates, sequence numbers, counts, and sums.
//
// Create implementations are free to modify computable fields in a file and should
// call the Batch's Validate function at the end of their execution.
func (batch *BatchMTE) Create() error {
	// generates sequence numbers and batch control
	if err := batch.build(); err != nil {
		return err
	}
	// Additional steps specific to batch type
	// ...

	if err := batch.Validate(); err != nil {
		return err
	}
	return nil
}
```

Finally add the batch type to the NewBatch factory in batch.go.

```go
//...
case MTE:
		return NewBatchMTE(bh), nil
//...
```

In order for the code to be merged with a Pull requests we require a `batchMTE_test.go` test file that covers the logic of the type. Refer to the [Go blog post on code coverage metrics](https://blog.golang.org/cover).

## References

* [Wikipedia: Automated Clearing House](https://en.wikipedia.org/wiki/Automated_Clearing_House)
* [NACHA ACH Network: How it Works](https://www.nacha.org/ach-network)
* [Federal ACH Directory](https://www.frbservices.org/EPaymentsDirectory/search.html)

## Format Specification

* [NACHA ACH File Details](https://achdevguide.nacha.org/ach-file-details)
* [Thomson Reuters ACH File Structure](https://www.thomsonreuters.com/en-us/help/accounting-cs/direct-deposit/ach-structure-and-contents.html)
* [Gusto: How ACH Works: A developer perspective](https://medium.com/gusto-engineering/how-ach-works-a-developer-perspective-part-5-1d998bbcd82c)

![ACH File Layout](https://github.com/moov-io/ach/blob/master/docs/ach_file_structure_shg.gif)

## Inspiration

* [ACH:Builder - Tools for Building ACH](http://search.cpan.org/~tkeefer/ACH-Builder-0.03/lib/ACH/Builder.pm)
* [mosscode / ach](https://github.com/mosscode/ach)
* [Helper for building ACH files in Ruby](https://github.com/jm81/ach)
* [Glenselle / nACH2](https://github.com/glenselle/nACH2)
