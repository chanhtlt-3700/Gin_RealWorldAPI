package models

// All trả về danh sách model cần AutoMigrate.
// Mỗi giai đoạn chỉ cần thêm model mới vào đây (Follow, Article, Comment, Tag...).
func All() []any {
	return []any{
		&User{},
	}
}
