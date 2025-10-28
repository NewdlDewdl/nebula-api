package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/UTDNebula/nebula-api/api/configs"

	"github.com/UTDNebula/nebula-api/api/schema"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var courseCollection *mongo.Collection = configs.GetCollection("courses")

// @Id				courseSearch
// @Router			/course [get]
// @Tags			Courses
// @Description	"Returns paginated list of courses matching the query's string-typed key-value pairs. See offset for more details on pagination."
// @Produce		json
// @Param			offset					query		number								false	"The starting position of the current page of courses (e.g. For starting at the 17th course, offset=16)."
// @Param			course_number			query		string								false	"The course's official number"
// @Param			subject_prefix			query		string								false	"The course's subject prefix"
// @Param			title					query		string								false	"The course's title"
// @Param			description				query		string								false	"The course's description"
// @Param			school					query		string								false	"The course's school"
// @Param			credit_hours			query		string								false	"The number of credit hours awarded by successful completion of the course"
// @Param			class_level				query		string								false	"The level of education that this course course corresponds to"
// @Param			activity_type			query		string								false	"The type of class this course corresponds to"
// @Param			grading					query		string								false	"The grading status of this course"
// @Param			internal_course_number	query		string								false	"The internal (university) number used to reference this course"
// @Param			lecture_contact_hours	query		string								false	"The weekly contact hours in lecture for a course"
// @Param			offering_frequency		query		string								false	"The frequency of offering a course"
// @Success		200						{object}	schema.APIResponse[[]schema.Course]	"A list of courses"
// @Failure		500						{object}	schema.APIResponse[string]			"A string describing the error"
// @Failure		400						{object}	schema.APIResponse[string]			"A string describing the error"
func CourseSearch(c *gin.Context) {
	var courses []schema.Course

	query, options, err := getQueryWithPagination[schema.Course]("Search", c)
	if err != nil {
		return
	}

	if err := executeFind(c, courseCollection, query, options, &courses); err != nil {
		return
	}

	respond(c, http.StatusOK, "success", courses)
}

// @Id				courseById
// @Router			/course/{id} [get]
// @Tags			Courses
// @Description	"Returns the course with given ID"
// @Produce		json
// @Param			id	path		string								true	"ID of the course to get"
// @Success		200	{object}	schema.APIResponse[schema.Course]	"A course"
// @Failure		500	{object}	schema.APIResponse[string]			"A string describing the error"
func CourseById(c *gin.Context) {
	var course schema.Course

	query, err := getQuery[schema.Course]("ById", c)
	if err != nil {
		return
	}

	if err := executeFindOne(c, courseCollection, query, &course); err != nil {
		return
	}

	respond(c, http.StatusOK, "success", course)
}

// @Id				courseAll
// @Router			/course/all [get]
// @Tags			Courses
// @Description	"Returns all courses"
// @Produce		json
// @Success		200	{object}	schema.APIResponse[[]schema.Course]	"All courses"
// @Failure		500	{object}	schema.APIResponse[string]			"A string describing the error"
func CourseAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var courses []schema.Course

	cursor, err := courseCollection.Find(ctx, bson.M{})
	if err != nil {
		respondWithInternalError(c, err)
		return
	}

	if err = cursor.All(ctx, &courses); err != nil {
		respondWithInternalError(c, err)
		return
	}

	respond(c, http.StatusOK, "success", courses)
}

// @Id				courseSectionSearch
// @Router			/course/sections [get]
// @Tags			Courses
// @Description	"Returns paginated list of sections of all the courses matching the query's string-typed key-value pairs. See former_offset and latter_offset for pagination details."
// @Produce		json
// @Param			former_offset			query		number									false	"The starting position of the current page of courses (e.g. For starting at the 17th course, former_offset=16)."
// @Param			latter_offset			query		number									false	"The starting position of the current page of sections (e.g. For starting at the 4th section, latter_offset=3)."
// @Param			course_number			query		string									false	"The course's official number"
// @Param			subject_prefix			query		string									false	"The course's subject prefix"
// @Param			title					query		string									false	"The course's title"
// @Param			description				query		string									false	"The course's description"
// @Param			school					query		string									false	"The course's school"
// @Param			credit_hours			query		string									false	"The number of credit hours awarded by successful completion of the course"
// @Param			class_level				query		string									false	"The level of education that this course course corresponds to"
// @Param			activity_type			query		string									false	"The type of class this course corresponds to"
// @Param			grading					query		string									false	"The grading status of this course"
// @Param			internal_course_number	query		string									false	"The internal (university) number used to reference this course"
// @Param			lecture_contact_hours	query		string									false	"The weekly contact hours in lecture for a course"
// @Param			offering_frequency		query		string									false	"The frequency of offering a course"
// @Success		200						{object}	schema.APIResponse[[]schema.Section]	"A list of sections"
// @Failure		500						{object}	schema.APIResponse[string]				"A string describing the error"
// @Failure		400						{object}	schema.APIResponse[string]				"A string describing the error"
func CourseSectionSearch() gin.HandlerFunc {
	return makeHandler("Search", courseSection)
}

// @Id				courseSectionById
// @Router			/course/{id}/sections [get]
// @Tags			Courses
// @Description	"Returns the all of the sections of the course with given ID"
// @Produce		json
// @Param			id	path		string									true	"ID of the course to get"
// @Success		200	{object}	schema.APIResponse[[]schema.Section]	"A list of sections"
// @Failure		500	{object}	schema.APIResponse[string]				"A string describing the error"
// @Failure		400	{object}	schema.APIResponse[string]				"A string describing the error"
func CourseSectionById() gin.HandlerFunc {
	return makeHandler("ById", courseSection)
}

// get the sections of the courses, filters depending on the flag
func courseSection(flag string, c *gin.Context) {
	var courseSections []schema.Section
	var courseQuery bson.M
	var err error

	courseQuery, err = getQuery[schema.Course](flag, c)
	if err != nil {
		return
	}

	paginateMap, err := configs.GetAggregateLimit(&courseQuery, c)
	if err != nil {
		respond(c, http.StatusBadRequest, "Error offset is not type integer", err.Error())
		return
	}

	courseSectionPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: courseQuery}},
		bson.D{{Key: "$skip", Value: paginateMap["former_offset"]}},
		bson.D{{Key: "$limit", Value: paginateMap["limit"]}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "sections"},
			{Key: "localField", Value: "sections"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sections"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$sections"},
			{Key: "preserveNullAndEmptyArrays", Value: false},
		}}},
		bson.D{{Key: "$replaceWith", Value: "$sections"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		bson.D{{Key: "$skip", Value: paginateMap["latter_offset"]}},
		bson.D{{Key: "$limit", Value: paginateMap["limit"]}},
	}

	if err := executeAggregate(c, courseCollection, courseSectionPipeline, &courseSections); err != nil {
		return
	}

	respond(c, http.StatusOK, "success", courseSections)
}

// @Id				courseProfessorSearch
// @Router			/course/professors [get]
// @Tags			Courses
// @Description	"Returns paginated list of professors of all the courses matching the query's string-typed key-value pairs. See former_offset and latter_offset for pagination details."
// @Produce		json
// @Param			former_offset			query		number									false	"The starting position of the current page of courses (e.g. For starting at the 17th course, former_offset=16)."
// @Param			latter_offset			query		number									false	"The starting position of the current page of professors (e.g. For starting at the 4th professor, latter_offset=3)."
// @Param			course_number			query		string									false	"The course's official number"
// @Param			subject_prefix			query		string									false	"The course's subject prefix"
// @Param			title					query		string									false	"The course's title"
// @Param			description				query		string									false	"The course's description"
// @Param			school					query		string									false	"The course's school"
// @Param			credit_hours			query		string									false	"The number of credit hours awarded by successful completion of the course"
// @Param			class_level				query		string									false	"The level of education that this course course corresponds to"
// @Param			activity_type			query		string									false	"The type of class this course corresponds to"
// @Param			grading					query		string									false	"The grading status of this course"
// @Param			internal_course_number	query		string									false	"The internal (university) number used to reference this course"
// @Param			lecture_contact_hours	query		string									false	"The weekly contact hours in lecture for a course"
// @Param			offering_frequency		query		string									false	"The frequency of offering a course"
// @Success		200						{object}	schema.APIResponse[[]schema.Professor]	"A list of professors"
// @Failure		500						{object}	schema.APIResponse[string]				"A string describing the error"
// @Failure		400						{object}	schema.APIResponse[string]				"A string describing the error"
func CourseProfessorSearch(c *gin.Context) {
	courseProfessor("Search", c)
}

// @Id				courseProfessorById
// @Router			/course/{id}/professors [get]
// @Tags			Courses
// @Description	"Returns the all of the professors of the course with given ID"
// @Produce		json
// @Param			id	path		string									true	"ID of the course to get"
// @Success		200	{object}	schema.APIResponse[[]schema.Professor]	"A list of professors"
// @Failure		500	{object}	schema.APIResponse[string]				"A string describing the error"
// @Failure		400	{object}	schema.APIResponse[string]				"A string describing the error"
func CourseProfessorById(c *gin.Context) {
	courseProfessor("ById", c)
}

// Get the professors of the courses, filters depending on the flag
func courseProfessor(flag string, c *gin.Context) {
	var courseProfessors []schema.Professor
	var courseQuery bson.M
	var err error

	if courseQuery, err = getQuery[schema.Course](flag, c); err != nil {
		return
	}

	paginateMap, err := configs.GetAggregateLimit(&courseQuery, c)
	if err != nil {
		respond(c, http.StatusBadRequest, "Error offset is not type integer", err.Error())
		return
	}

	courseProfessorPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: courseQuery}},
		bson.D{{Key: "$skip", Value: paginateMap["former_offset"]}},
		bson.D{{Key: "$limit", Value: paginateMap["limit"]}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "sections"},
			{Key: "localField", Value: "sections"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sections"},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "professors"},
			{Key: "localField", Value: "sections.professors"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "professors"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$professors"},
			{Key: "preserveNullAndEmptyArrays", Value: false},
		}}},
		bson.D{{Key: "$replaceWith", Value: "$professors"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		bson.D{{Key: "$skip", Value: paginateMap["latter_offset"]}},
		bson.D{{Key: "$limit", Value: paginateMap["limit"]}},
	}

	if err := executeAggregate(c, courseCollection, courseProfessorPipeline, &courseProfessors); err != nil {
		return
	}

	respond(c, http.StatusOK, "success", courseProfessors)
}

// @Id				trendsCourseSectionSearch
// @Router			/course/sections/trends [get]
// @Tags			Courses
// @Description	"Returns all of the given course's sections with Course and Professor data embedded. Specialized high-speed convenience endpoint for UTD Trends internal use; limited query flexibility."
// @Produce		json
// @Param			course_number	query		string									true	"The course's official number"
// @Param			subject_prefix	query		string									true	"The course's subject prefix"
// @Success		200				{object}	schema.APIResponse[[]schema.Section]	"A list of Sections"
// @Failure		500				{object}	schema.APIResponse[string]				"A string describing the error"
func TrendsCourseSectionSearch(c *gin.Context) {
	var courseSections []schema.Section
	courseQuery := bson.M{"_id": c.Query("subject_prefix") + c.Query("course_number")}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: courseQuery}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$sections"},
			{Key: "preserveNullAndEmptyArrays", Value: false},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "professors"},
			{Key: "localField", Value: "sections.professors"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sections.professor_details"},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "courses"},
			{Key: "localField", Value: "sections.course_reference"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sections.course_details"},
		}}},
		bson.D{{Key: "$replaceWith", Value: "$sections"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	trendsCollection := configs.GetCollection("trends_course_sections")
	if err := executeAggregate(c, trendsCollection, pipeline, &courseSections); err != nil {
		return
	}

	respond(c, http.StatusOK, "success", courseSections)
}
