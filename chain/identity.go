package chain

// Identity returns v unchanged.
//
// It exists for the deduplication methods. [Seq.UniqBy] and [Slice.UniqBy] take a
// key function, and there is no zero-argument Uniq because a method cannot add
// the comparable constraint its receiver lacks. Pass Identity to deduplicate on
// the values themselves:
//
//	chain.OfSlice(ids).UniqBy(chain.Identity)
//
// T is comparable because that is the only context Identity is useful in: the key
// a UniqBy derives has to be comparable.
func Identity[T comparable](v T) T {
	return v
}
