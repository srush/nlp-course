package nlp

type State int;
type Outcome int;

type HMM struct {
	num_states int
	num_outcomes int
	transitions []Multinomial
	emissions []Multinomial
	start Multinomial
}

type HMMCounts struct {
	num_states int
	num_outcomes int
	transitions []MultinomialCounts
	emissions []MultinomialCounts
	start MultinomialCounts
}

func MaximumLikelihoodHMM(counts HMMCounts) HMM {
	hmm := HMM {
		num_states : counts.num_states,
		num_outcomes : counts.num_outcomes,
		emissions: make([]Multinomial, counts.num_states),
		transitions: make([]Multinomial, counts.num_states),
		start : counts.start.MaximumLikelihood(),
	}
	for state := 0; state < hmm.num_states; state++  {
		hmm.emissions[state] = counts.emissions[state].MaximumLikelihood()
		hmm.transitions[state] = counts.transitions[state].MaximumLikelihood()
	}
	hmm.start = counts.start.MaximumLikelihood()
	return hmm
}

func NewHMMCounts(num_states int, num_outcomes int) HMMCounts {
	counts := HMMCounts {
		num_states: num_states,
		num_outcomes: num_outcomes,
		emissions : make([]MultinomialCounts, num_states),
		transitions : make([]MultinomialCounts, num_states),
		start : NewCounts(),
	}
	for emission := 0; emission < counts.num_states; emission++ {
		counts.emissions[emission] = NewCounts()
	}
	for state := 0; state < counts.num_states; state++ {
		counts.transitions[state] = NewCounts()
	}
	return counts
}

func (counts *HMMCounts) IncStart(state State) {
	counts.start.Inc(int(state))
}

func (counts *HMMCounts) IncTransition(state State, next_state State) {
	counts.transitions[state].Inc(int(next_state))
}

func (counts *HMMCounts) IncEmission(state State, outcome Outcome) {
	counts.emissions[state].Inc(int(outcome))
}

func (hmm HMM) ProbStart(state State) float64 {
	return hmm.start.Prob(int(state))
}

func (hmm HMM) ProbTransition(state State, next_state State) float64 {
	return hmm.transitions[state].Prob(int(next_state))
}

func (hmm HMM) ProbEmission(state State, outcome Outcome) float64 {
	return hmm.emissions[state].Prob(int(outcome))
}

type loc struct {
	position int
	state State 
}

type chartType struct {
	cells map[loc] cell
}

type cell struct {
	score float64
	best State
}

func (hmm HMM) RunViterbi(outcomes []Outcome) (float64, []State) {
	chart := &chartType {
		cells : map[loc]cell {}, 
	}

	// Initialize
	for state := 0; state < hmm.num_states; state++ {
		chart.cells[loc{0, State(state)}].score = hmm.ProbStart(State(state)) 
	}

	// Main loop.
	for position, outcome := range outcomes {
		if position == 0 { continue }
		for  state State = 0; state < hmm.num_states; state++ {
			for prev_state := 0; prev_state < hmm.num_states; prev_state++ {
				score := chart.cells[loc{position - 1, State(prev_state)}].score + hmm.ProbTransition(State(prev_state), State(state)) 
				if score > chart.cells[loc{position, State(state)}].score {
					chart.cells[loc{position, State(state)}] = cell{ score, State(prev_state) }
				}
			}
			chart.cells[loc{position, State(state)}].score += hmm.ProbEmission(State(state), State(outcome)) 
		}
	}

	// Find best state. 
	var best int
	best_score := 0.0 
	var states []State 
	for state := 0; state < hmm.num_states; state++ {
		if s:= chart.cells[loc{end, State(state)}].score; s > best_score {
			best_score = s
			best = state
		}
	}

	// Follow back pointers to find the final path.
	cur_state = best
	for position = end; position >=0; position++ {
		states = append(states, state)
		cur_state = chart.cells[loc{position, State(cur_state)}].best
	}
	return best_score, states
}