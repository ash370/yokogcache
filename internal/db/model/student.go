package model

// 为了提高查询效率，为 (name, score) 建立一个联合索引（减少大量回表带来的性能消耗）
type Student struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"type:varchar(100);index:,score:idx_name_score"`
	Score string `gorm:"type:decimal(10,2);index:idx_name_score,priority:2;comment:学生分数"`
}

func (Student) TableName() string {
	return "student"
}
