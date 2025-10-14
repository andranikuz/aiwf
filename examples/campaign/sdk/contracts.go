package blog

// NOTE: Ниже показаны заглушки типов. Замените `map[string]any` на структуры,
//      соответствующие JSON Schema, чтобы получить типизированный SDK.
// ContentStrategyInput соответствует схеме strategy_input.json.
type ContentStrategyInput struct {
    BudgetLevel string `json:"budget_level"`
    PrimaryGoal string `json:"primary_goal"`
    ValueMap []map[string]any `json:"value_map"`
}
// ContentStrategyOutput соответствует схеме strategy_output.json.
type ContentStrategyOutput struct {
    BudgetNotes string `json:"budget_notes"`
    Channels []map[string]any `json:"channels"`
}


// LaunchTimelineInput соответствует схеме timeline_input.json.
type LaunchTimelineInput struct {
    BudgetNotes string `json:"budget_notes"`
    Channels []map[string]any `json:"channels"`
}
// LaunchTimelineOutput соответствует схеме timeline_output.json.
type LaunchTimelineOutput struct {
    Milestones []map[string]any `json:"milestones"`
}


// MarketResearchInput соответствует схеме research_input.json.
type MarketResearchInput struct {
    Description string `json:"description"`
    Goals []string `json:"goals"`
    ProductName string `json:"product_name"`
    TargetMarket string `json:"target_market"`
}
// MarketResearchOutput соответствует схеме research_output.json.
type MarketResearchOutput struct {
    Segments []map[string]any `json:"segments"`
    Summary string `json:"summary"`
}


// RiskAssessmentInput соответствует схеме risk_input.json.
type RiskAssessmentInput struct {
    Milestones []map[string]any `json:"milestones"`
}
// RiskAssessmentOutput соответствует схеме risk_output.json.
type RiskAssessmentOutput struct {
    Risks []map[string]any `json:"risks"`
}


// ValuePropositionInput соответствует схеме proposition_input.json.
type ValuePropositionInput struct {
    ProductName string `json:"product_name"`
    Segments []map[string]any `json:"segments"`
}
// ValuePropositionOutput соответствует схеме proposition_output.json.
type ValuePropositionOutput struct {
    Narrative string `json:"narrative"`
    ValueMap []map[string]any `json:"value_map"`
}


// CampaignLaunchInput описывает вход workflow.
type CampaignLaunchInput = MarketResearchInput

// CampaignLaunchOutput описывает выход workflow.
type CampaignLaunchOutput = RiskAssessmentOutput

