package parameterdata

type Topic struct {
	Name     string   `json:"Name"`
	Children []*Topic `json:"Children,omitempty"`
}

var MathTopicsList = []*Topic{
	{
		Name: "Algebra",
		Children: []*Topic{
			{Name: "Linear equations in 1 variable"},
			{Name: "Linear equations in 2 variables"},
			{Name: "Linear functions"},
			{Name: "Systems of 2 linear equations in 2 variables"},
			{Name: "Linear inequalities in 1 or 2 variables"},
		},
	},
	{
		Name: "Advanced math",
		Children: []*Topic{
			{Name: "Equivalent expressions"},
			{Name: "Nonlinear equations in 1 variable"},
			{Name: "Systems of equations in 2 variables"},
			{Name: "Nonlinear functions"},
		},
	},
	{
		Name: "Problem solving and data analysis",
		Children: []*Topic{
			{Name: "Ratios, rates, proportional relationships, and units"},
			{Name: "Percentages"},
			{Name: "One-variable data: distributions and measures of center and spread"},
			{Name: "Two-variable data: models and scatterplots"},
			{Name: "Probability and conditional probability"},
			{Name: "Inference from sample statistics and margin of error"},
			{Name: "Evaluating statistical claims: observational studies and experiments"},
		},
	},
	{
		Name: "Geometry and trigonometry",
		Children: []*Topic{
			{Name: "Area and volume formulas"},
			{Name: "Lines, angles, and triangles"},
			{Name: "Right triangles and trigonometry"},
			{Name: "Circles"},
		},
	},
}

var ReadingTopicsList = []*Topic{
	{
		Name: "Information and ideas",
		Children: []*Topic{
			{Name: "Information and ideas"},
		},
	},
	{
		Name: "Craft and structure",
		Children: []*Topic{
			{Name: "Craft and structure"},
		},
	},
	{
		Name: "Expression of ideas",
		Children: []*Topic{
			{Name: "Expression of ideas"},
		},
	},
	{
		Name: "Standard English conventions",
		Children: []*Topic{
			{Name: "Standard English conventions"},
		},
	},
}

type LessonModule struct {
	Name     string   `json:"Name"`
	VideoIDs []string `json:"VideoIDs"`
}

var LessonModules = []*LessonModule{
	{
		Name: "Linear equations in 1 variable",
		VideoIDs: []string{
			"65e68427f5ad8e134d716e8f",
			"65e68f1ac715966a3872061e",
		},
	},
	{
		Name: "Linear equations in 2 variables",
		VideoIDs: []string{
			"65e68427f5ad8e134d716e8f",
			"65e68f1ac715966a3872061e",
		},
	},
}

type PracticeModule struct {
	Name        string   `json:"Name"`
	QuestionIDs []string `json:"QuestionIDs"`
}

var PracticeModules = []*PracticeModule{
	{
		Name: "Linear equations in 1 variable",
		QuestionIDs: []string{
			"65bad3c908992ac645d86bc5",
			"65bae13d08992ac645d86bc6",
			"65bc543a08992ac645d86bed",
			"65bb098408992ac645d86bcd",
			"65bb1e5108992ac645d86bda",
		},
	},
	{
		Name: "Linear equations in 2 variables",
		QuestionIDs: []string{
			"65bad3c908992ac645d86bc5",
			"65bae13d08992ac645d86bc6",
			"65bc543a08992ac645d86bed",
			"65bb098408992ac645d86bcd",
			"65bb1e5108992ac645d86bda",
		},
	},
}

type TestRepresentation struct {
	Name          string     `json:"Name"`
	QuestionLists [][]string `json:"QuestionLists"`
}

var Tests = []*TestRepresentation{
	{
		Name: "Practice test 1",
		QuestionLists: [][]string{
			{"65bad3c908992ac645d86bc5", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bae26b08992ac645d86bc7", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bae36808992ac645d86bc8", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bae54a08992ac645d86bc9", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
		},
	},
	{
		Name: "Practice test 2",
		QuestionLists: [][]string{
			{"65bad3c908992ac645d86bc5", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bad3c908992ac645d86bc5", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bad3c908992ac645d86bc5", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
			{"65bad3c908992ac645d86bc5", "65bae13d08992ac645d86bc6", "65bc543a08992ac645d86bed", "65bb098408992ac645d86bcd", "65bb1e5108992ac645d86bda"},
		},
	},
}
