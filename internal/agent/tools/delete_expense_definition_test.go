package tools_test

func (s *Suite) TestDeleteExpenseDefinition_Schema() {
	definition := s.deleteExpense.Definition()

	s.Equal("delete_expense", definition.Name)
	s.NotEmpty(definition.Description)
	s.Equal([]any{"id"}, definition.Parameters["required"])
	properties := definition.Parameters["properties"].(map[string]any)
	s.ElementsMatch([]string{"id"}, keys(properties))
}
