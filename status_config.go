package service

type StatusConfig struct {
	NotFound        int  `mapstructure:"not_found" json:"notFound" gorm:"column:notfound" bson:"notFound" dynamodbav:"notFound" firestore:"notFound"`
	Success         int  `mapstructure:"success" json:"success" gorm:"column:success" bson:"success" dynamodbav:"success" firestore:"success"`
	VersionError    int  `mapstructure:"version_error" json:"versionError" gorm:"column:versionerror" bson:"versionError" dynamodbav:"versionError" firestore:"versionError"`
	Error           int  `mapstructure:"error" json:"error" gorm:"column:error" bson:"error" dynamodbav:"error" firestore:"error"`
}
