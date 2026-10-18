package tools_test

func (s *Suite) TestListExpensesDefinition_Schema() {
	definition := s.listExpenses.Definition()

	s.Equal("list_expenses", definition.Name)
	s.NotEmpty(definition.Description)
	s.Equal("object", definition.Parameters["type"])
	s.NotContains(definition.Parameters, "required")
	properties := definition.Parameters["properties"].(map[string]any)
	s.ElementsMatch([]string{"from", "to", "category"}, keys(properties))
	s.Equal(categoryCodes(), properties["category"].(map[string]any)["enum"])
	for _, name := range userProperties {
		s.NotContains(properties, name)
	}
}
