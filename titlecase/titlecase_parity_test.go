package titlecase

import "testing"

// Upstream parity tests for blakeembrey/change-case, package title-case.
//
// Every vector below is copied verbatim from the upstream test suite:
//
//	https://raw.githubusercontent.com/blakeembrey/change-case/main/packages/title-case/src/index.spec.ts
//
// (which in turn derives from https://github.com/gouch/to-title-case). The
// upstream implementation this port targets is:
//
//	https://raw.githubusercontent.com/blakeembrey/change-case/main/packages/title-case/src/index.ts
//
// Option-bearing upstream cases ({ sentenceCase: true } and
// { smallWords: new Set() }) are mapped onto this port's Options type.

func TestParityTitleCase(t *testing.T) {
	cases := []struct {
		in   string
		want string
		opt  *Options
	}{
		{in: "one two", want: "One Two"},
		{in: "one two three", want: "One Two Three"},
		{
			in:   "Start a an and as at but by en for if in nor of on or per the to v vs via end",
			want: "Start a an and as at but by en for if in nor of on or per the to v vs via End",
		},
		{in: "a small word starts", want: "A Small Word Starts"},
		{in: "small word ends on", want: "Small Word Ends On"},
		{in: "questions?", want: "Questions?"},
		{in: "Two questions?", want: "Two Questions?"},
		{in: "one sentence. two sentences.", want: "One Sentence. Two Sentences."},
		{in: "we keep NASA capitalized", want: "We Keep NASA Capitalized"},
		{in: "pass camelCase through", want: "Pass camelCase Through"},
		{in: "this sub-phrase is nice", want: "This Sub-Phrase Is Nice"},
		{in: "follow step-by-step instructions", want: "Follow Step-by-Step Instructions"},
		{in: "easy as one-two-three end", want: "Easy as One-Two-Three End"},
		{in: "start on-demand end", want: "Start On-Demand End"},
		{in: "start in-or-out end", want: "Start In-or-Out End"},
		{in: "start e-commerce end", want: "Start E-Commerce End"},
		{in: "start e-mail end", want: "Start E-Mail End"},
		{in: "your hair[cut] looks (nice)", want: "Your Hair[cut] Looks (Nice)"},
		{in: "keep that colo(u)r", want: "Keep that Colo(u)r"},
		{in: "leave Q&A unscathed", want: "Leave Q&A Unscathed"},
		{in: "piña colada while you listen to ænima", want: "Piña Colada While You Listen to Ænima"},
		{in: "start title – end title", want: "Start Title – End Title"},
		{in: "start title–end title", want: "Start Title–End Title"},
		{in: "start title — end title", want: "Start Title — End Title"},
		{in: "start title—end title", want: "Start Title—End Title"},
		{in: "start title - end title", want: "Start Title - End Title"},
		{in: "don't break", want: "Don't Break"},
		{in: `"double quotes"`, want: `"Double Quotes"`},
		{in: `double quotes "inner" word`, want: `Double Quotes "Inner" Word`},
		{in: "fancy double quotes “inner” word", want: "Fancy Double Quotes “Inner” Word"},
		{in: "'single quotes'", want: "'Single Quotes'"},
		{in: "single quotes 'inner' word", want: "Single Quotes 'Inner' Word"},
		{in: "fancy single quotes ‘inner’ word", want: "Fancy Single Quotes ‘Inner’ Word"},
		{in: "“‘a twice quoted subtitle’”", want: "“‘A Twice Quoted Subtitle’”"},
		{in: "have you read “The Lottery”?", want: "Have You Read “The Lottery”?"},
		{in: "one: two", want: "One: Two"},
		{in: "one two: three four", want: "One Two: Three Four"},
		{in: `one two: "Three Four"`, want: `One Two: "Three Four"`},
		{in: "one on: an end", want: "One On: An End"},
		{in: `one on: "an end"`, want: `One On: "An End"`},
		{in: "email email@example.com address", want: "Email email@example.com Address"},
		{in: "you have an https://example.com/ title", want: "You Have an https://example.com/ Title"},
		{in: "_underscores around words_", want: "_Underscores Around Words_"},
		{in: "*asterisks around words*", want: "*Asterisks Around Words*"},
		{in: "this vs that", want: "This vs That"},
		{in: "this *vs* that", want: "This *vs* That"},
		{in: "this v that", want: "This v That"},
		{in: "this vs. that", want: "This Vs. That"},
		{in: "this v. that", want: "This V. That"},
		{in: "", want: ""},
		{
			in:   "Scott Moritz and TheStreet.com’s million iPhone la-la land",
			want: "Scott Moritz and TheStreet.com’s Million iPhone La-La Land",
		},
		{
			in:   "Notes and observations regarding Apple’s announcements from ‘The Beat Goes On’ special event",
			want: "Notes and Observations Regarding Apple’s Announcements From ‘The Beat Goes On’ Special Event",
		},
		{in: "2018", want: "2018"},
		{
			in:   "the quick brown fox jumps over the lazy dog",
			want: "The Quick Brown Fox Jumps over the Lazy Dog",
		},
		{in: "newcastle upon tyne", want: "Newcastle upon Tyne"},
		{in: "newcastle *upon* tyne", want: "Newcastle *upon* Tyne"},
		{
			in:   "Is human activity responsible for the climate emergency? New report calls it ‘unequivocal.’",
			want: "Is Human Activity Responsible for the Climate Emergency? New Report Calls It ‘Unequivocal.’",
		},
		{in: "лев николаевич толстой", want: "Лев Николаевич Толстой"},
		{in: "Read foo-bar.com", want: "Read foo-bar.com"},
		{in: "cowboy bebop: the movie", want: "Cowboy Bebop: The Movie"},
		{in: "a thing. the thing. and more.", want: "A Thing. The Thing. And More."},
		{in: `"a quote." a test.`, want: `"A Quote." A Test.`},
		{in: `"The U.N." a quote.`, want: `"The U.N." A Quote.`},
		{in: `"The U.N.". a quote.`, want: `"The U.N.". A Quote.`},
		{in: `"The U.N.". a quote.`, want: `"The U.N.". A quote.`, opt: &Options{SentenceCase: true}},
		{in: `"go without"`, want: `"Go Without"`},
		{in: "the iPhone: a quote", want: "The iPhone: A Quote"},
		{in: "the iPhone: a quote", want: "The iPhone: a quote", opt: &Options{SentenceCase: true}},
		{in: "the U.N. and me", want: "The U.N. and Me"},
		{in: "the *U.N.* and me", want: "The *U.N.* and Me"},
		{in: "the U.N. and me", want: "The U.N. and me", opt: &Options{SentenceCase: true}},
		{in: "the U.N. and me", want: "The U.N. And Me", opt: &Options{SmallWords: map[string]bool{}}},
		{in: "start-and-end", want: "Start-and-End"},
		{in: "go-to-iPhone", want: "Go-to-iPhone"},
		{in: "the go-to", want: "The Go-To"},
		{in: "the go-to", want: "The go-to", opt: &Options{SentenceCase: true}},
		{in: "this to-go", want: "This To-Go"},
		{in: "test(ing)", want: "Test(ing)"},
		{in: "test(s)", want: "Test(s)"},
		{in: "Keep #tag", want: "Keep #tag"},
		{in: `"Hello world", says John.`, want: `"Hello World", Says John.`},
		{in: `"Hello world", says John.`, want: `"Hello world", says John.`, opt: &Options{SentenceCase: true}},
		{in: "foo/bar", want: "Foo/Bar"},
		{in: "this is the *end.*", want: "This Is the *End.*"},
		{in: "*something about me?* and you.", want: "*Something About Me?* And You."},
		{in: "*something about me?* and you.", want: "*Something about me?* And you.", opt: &Options{SentenceCase: true}},
		{in: "something about _me-too?_ and you.", want: "Something About _Me-Too?_ And You."},
		{in: "something about _me_? and you.", want: "Something About _Me_? And You."},
		{in: "something about _me_? and you.", want: "Something about _me_? And you.", opt: &Options{SentenceCase: true}},
		{in: "something about _me-too_? and you too.", want: "Something About _Me-Too_? And You Too."},
		{in: "an example. i.e. test.", want: "An Example. I.e. Test."},
		{in: "an example, i.e. test.", want: "An Example, I.e. Test."},
		{in: `an example. "i.e. test."`, want: `An Example. "I.e. Test."`},
		{in: "an example. i.e. test.", want: "An example. I.e. test.", opt: &Options{SentenceCase: true}},
		{in: "an example, i.e. test.", want: "An example, i.e. test.", opt: &Options{SentenceCase: true}},
		{in: `an example. "i.e. test."`, want: `An example. "I.e. test."`, opt: &Options{SentenceCase: true}},
		{in: "friday the 13th", want: "Friday the 13th"},
		{in: "21st century", want: "21st Century"},
		{in: "foo\nbar", want: "Foo\nBar"},
		{in: "foo\nbar\nbaz", want: "Foo\nBar\nBaz"},
		{in: "friday\nthe 13th", want: "Friday\nThe 13th"},
	}

	for _, c := range cases {
		var got string
		if c.opt != nil {
			got = TitleCase(c.in, *c.opt)
		} else {
			got = TitleCase(c.in)
		}
		if got != c.want {
			t.Errorf("TitleCase(%q, %+v) = %q, want %q", c.in, c.opt, got, c.want)
		}
	}
}
