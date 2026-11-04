package models

// AcceptanceCriterionModel representa la tabla "acceptance_criterion" en PostgreSQL.
type AcceptanceCriterionModel struct {
	ID          uint          `gorm:"column:id;primaryKey;autoIncrement"`
	HistoriaID  uint          `gorm:"column:backlog_item_id;not null"`
	Descripcion string        `gorm:"column:description;type:text;not null"`
	Position    int           `gorm:"column:position;not null;default:1"`
	IsMet       bool          `gorm:"column:is_met;not null;default:false"`
	Historia    HistoriaModel `gorm:"foreignKey:HistoriaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (AcceptanceCriterionModel) TableName() string {
	return "acceptance_criterion"
}
