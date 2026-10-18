package entity

import "testing"

func TestCategoryValid(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		want     bool
	}{
		{name: "food", category: CategoryFood, want: true},
		{name: "transport", category: CategoryTransport, want: true},
		{name: "housing", category: CategoryHousing, want: true},
		{name: "entertainment", category: CategoryEntertainment, want: true},
		{name: "health", category: CategoryHealth, want: true},
		{name: "shopping", category: CategoryShopping, want: true},
		{name: "other", category: CategoryOther, want: true},
		{name: "unknown", category: "travel", want: false},
		{name: "russian name", category: "еда", want: false},
		{name: "wrong case", category: "Food", want: false},
		{name: "empty", category: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.Valid(); got != tt.want {
				t.Errorf("Category(%q).Valid() = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestCategoriesAreValidAndUnique(t *testing.T) {
	categories := Categories()
	if len(categories) != 7 {
		t.Fatalf("len(Categories()) = %d, want 7", len(categories))
	}
	seen := make(map[Category]bool, len(categories))
	for _, c := range categories {
		if !c.Valid() {
			t.Errorf("Categories() contains invalid category %q", c)
		}
		if seen[c] {
			t.Errorf("Categories() contains duplicate %q", c)
		}
		seen[c] = true
	}
}

func TestCategoriesReturnsCopy(t *testing.T) {
	first := Categories()
	first[0] = "mutated"
	if got := Categories()[0]; got != CategoryFood {
		t.Errorf("Categories()[0] = %q after mutating previous result, want %q", got, CategoryFood)
	}
}
