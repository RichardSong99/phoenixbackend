package datacube

import (
	"context"
	"example/goserver/parameterdata"
	"example/goserver/question"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataCubeService struct {
	collection      *mongo.Collection
	questionService *question.QuestionService
}

func NewDataCubeService(client *mongo.Client, questionService *question.QuestionService) *DataCubeService {
	collection := client.Database("test").Collection("datacubes")
	return &DataCubeService{
		collection:      collection,
		questionService: questionService,
	}
}
func (s *DataCubeService) GetDataCubeCollection() *mongo.Collection {
	return s.collection
}

func (s *DataCubeService) GetDataCube(userIDObj *primitive.ObjectID) (*DataCube, error) {
	var dataCube DataCube

	if userIDObj == nil {
		return nil, fmt.Errorf("userIDObj is nil")
	}

	// Create a filter to find the data cube for the given user
	filter := bson.M{"user_id": userIDObj}

	// Find the data cube
	err := s.collection.FindOne(context.TODO(), filter).Decode(&dataCube)
	if err != nil {

		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no data cube found for user")
		}
		return nil, fmt.Errorf("error getting data cube for user: %w", err)
	}

	return &dataCube, nil
}

func (s *DataCubeService) ComputeDataCube(userIDObj *primitive.ObjectID) (*DataCube, error) {
	fmt.Println("Computing data cube for user", userIDObj)
	// Create a context
	ctx := context.TODO()

	// Get the combined statistics
	combinedStats, err := s.questionService.GetCombinedCubeStatistics(ctx, userIDObj)
	if err != nil {
		return nil, fmt.Errorf("error getting combined statistics: %w", err)
	}

	// Compute the data cube using the combined statistics
	// For now, let's just create a new data cube and set its UserID field
	dataCube := &DataCube{
		UserID: *userIDObj,
		Rows:   make(map[string]Row),
	}

	for _, topicStat := range combinedStats {
		row := Row{Cells: make(map[string]Cell)}

		// Compute the data cube using the combined statistics
		// For now, let's just set the total to the total number of questions

		var allStatus = []string{"unattempted", "correct", "incorrect", "omitted"}

		// Initialize row.Cells with zero values for all statuses
		zero := 0.0
		for _, status := range allStatus {
			row.Cells[status] = Cell{
				Values: map[string]*float64{
					"easy":        &zero,
					"medium":      &zero,
					"hard":        &zero,
					"extreme":     &zero,
					"hardextreme": &zero,
					"total":       &zero,
				},
			}
		}

		if topicStat.Statuses != nil {
			for statusName, status := range topicStat.Statuses {
				// easy is equal to status.Difficulties.Easy if it exists, otherwise 0
				easy := 0.0
				if status.Difficulties.Easy != nil {
					easy = float64(*status.Difficulties.Easy)
				}

				// medium is equal to status.Difficulties.Medium if it exists, otherwise 0
				medium := 0.0
				if status.Difficulties.Medium != nil {
					medium = float64(*status.Difficulties.Medium)
				}

				// hard is equal to status.Difficulties.Hard if it exists, otherwise 0
				hard := 0.0
				if status.Difficulties.Hard != nil {
					hard = float64(*status.Difficulties.Hard)
				}

				// extreme is equal to status.Difficulties.Extreme if it exists, otherwise 0
				extreme := 0.0
				if status.Difficulties.Extreme != nil {
					extreme = float64(*status.Difficulties.Extreme)
				}

				// hardextreme is the sum of hard and extreme
				hardExtreme := hard + extreme

				// total is the sum of easy, medium, hard, and extreme
				total := easy + medium + hard + extreme

				// Add the cell to the row
				row.Cells[statusName] = Cell{
					Values: map[string]*float64{
						"easy":        &easy,
						"medium":      &medium,
						"hard":        &hard,
						"extreme":     &extreme,
						"hardextreme": &hardExtreme,
						"total":       &total,
					},
				}
			}
		}

		// Add the total cell to the row
		row.Cells["total"] = sumCells([]Cell{
			row.Cells["unattempted"], row.Cells["correct"], row.Cells["incorrect"], row.Cells["omitted"],
		})

		row.Cells["attempted"] = sumCells([]Cell{
			row.Cells["correct"], row.Cells["incorrect"], row.Cells["omitted"],
		})

		// Add the row to the data cube
		dataCube.Rows[topicStat.Topic] = row
	}

	// add subtotal rows for each topic with children
	mathTopicRows := []Row{}

	for _, topic := range parameterdata.MathTopicsList {

		subTopicRows := []Row{}

		for _, subtopic := range topic.Children {
			// access the row for the subtopic
			subtopicRow := dataCube.Rows[subtopic.Name]
			subTopicRows = append(subTopicRows, subtopicRow)
			// add the cells of the subtopic row to the cells of the row
		}

		// sum the rows for the subtopics
		summedRow := sumRows(subTopicRows)

		// add the summed row to the data cube
		dataCube.Rows[topic.Name] = summedRow
		mathTopicRows = append(mathTopicRows, summedRow)
	}

	// sum the rows for the math topics
	summedMathRow := sumRows(mathTopicRows)
	// add the summed row to the data cube
	dataCube.Rows["Math"] = summedMathRow

	//add subtotal rows for each topic with children, reading
	readingTopicRows := []Row{}
	for _, topic := range parameterdata.ReadingTopicsList {
		subTopicRows := []Row{}
		for _, subtopic := range topic.Children {
			// access the row for the subtopic
			subtopicRow := dataCube.Rows[subtopic.Name]
			subTopicRows = append(subTopicRows, subtopicRow)
			// add the cells of the subtopic row to the cells of the row
		}
		// sum the rows for the subtopics
		summedRow := sumRows(subTopicRows)
		// add the summed row to the data cube
		dataCube.Rows[topic.Name] = summedRow
		readingTopicRows = append(readingTopicRows, summedRow)
	}
	// sum the rows for the reading topics
	summedReadingRow := sumRows(readingTopicRows)
	// add the summed row to the data cube
	dataCube.Rows["Reading"] = summedReadingRow

	// add the Math and Reading rows for the "Total" row
	totalRow := sumRows([]Row{summedMathRow, summedReadingRow})
	dataCube.Rows["Total"] = totalRow

	// calculate usage and accuracy for all rows, add the cells...
	for _, row := range dataCube.Rows {
		// calculate usage
		usage := divideCells(row.Cells["attempted"], row.Cells["total"])
		row.Cells["usage"] = usage

		// calculate accuracy
		accuracy := divideCells(row.Cells["correct"], row.Cells["attempted"])
		row.Cells["accuracy"] = accuracy
	}

	// update the datacube in the database if it already exists, if not create a new one...
	_, err = s.collection.ReplaceOne(ctx, bson.M{"user_id": *userIDObj}, dataCube, options.Replace().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("error updating data cube: %w", err)
	}

	return dataCube, nil
}

func sumRows(rows []Row) Row {
	// create a new row
	summedRow := Row{Cells: make(map[string]Cell)}

	// create a map to hold the cells for each column
	columns := make(map[string][]Cell)

	// iterate through the rows
	for _, row := range rows {
		// add the cells of the row to the corresponding column
		for cellName, cell := range row.Cells {
			columns[cellName] = append(columns[cellName], cell)
		}
	}

	// sum the cells in each column and add them to the summed row
	for cellName, cells := range columns {
		summedRow.Cells[cellName] = sumCells(cells)
	}

	return summedRow
}

func sumCells(cells []Cell) Cell {
	// create a new cell
	summedCell := Cell{
		Values: make(map[string]*float64),
	}

	// iterate through the cells
	for _, cell := range cells {
		// add the values of the cells to the values of the summed cell
		for key, value := range cell.Values {
			if value != nil {
				if summedCell.Values[key] == nil {
					summedCell.Values[key] = new(float64)
				}
				*summedCell.Values[key] += *value
			}
		}
	}

	return summedCell
}

func divideCells(cell1 Cell, cell2 Cell) Cell {
	// create a new cell
	dividedCell := Cell{
		Values: make(map[string]*float64),
	}

	// iterate through the cells
	for key, value := range cell1.Values {
		if value != nil {
			if cell2.Values[key] != nil && *cell2.Values[key] != 0 {
				dividedCell.Values[key] = new(float64)
				*dividedCell.Values[key] = *value / *cell2.Values[key]
			} else {
				dividedCell.Values[key] = new(float64)
				*dividedCell.Values[key] = 0
			}
		}
	}

	return dividedCell
}
