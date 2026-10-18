package tools_test

func (s *Suite) TestAddExpenseDefinition_Schema() {
	definition := s.addExpense.Definition()

	s.Equal("add_expense", definition.Name)
	s.NotEmpty(definition.Description)
	s.Equal("object", definition.Parameters["type"])
	s.Equal([]any{"amount", "category", "comment"}, definition.Parameters["required"])
	properties := definition.Parameters["properties"].(map[string]any)
	s.ElementsMatch([]string{"amount", "category", "comment", "date"}, keys(properties))
	s.Equal(categoryCodes(), properties["category"].(map[string]any)["enum"])
	for _, name := range userProperties {
		s.NotContains(properties, name)
	}
}
