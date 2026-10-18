package tools_test

func (s *Suite) TestUpdateExpenseDefinition_Schema() {
	definition := s.updateExpense.Definition()

	s.Equal("update_expense", definition.Name)
	s.NotEmpty(definition.Description)
	s.Equal([]any{"id"}, definition.Parameters["required"])
	properties := definition.Parameters["properties"].(map[string]any)
	s.ElementsMatch([]string{"id", "amount", "category", "comment", "date"}, keys(properties))
	s.Equal(categoryCodes(), properties["category"].(map[string]any)["enum"])
	for _, name := range userProperties {
		s.NotContains(properties, name)
	}
}
