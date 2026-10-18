package tools_test

func (s *Suite) TestGetLastExpenseDefinition_Schema() {
	definition := s.getLastExpense.Definition()

	s.Equal("get_last_expense", definition.Name)
	s.NotEmpty(definition.Description)
	s.Equal("object", definition.Parameters["type"])
	s.Empty(definition.Parameters["properties"])
}
