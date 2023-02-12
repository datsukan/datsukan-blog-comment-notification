package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Input は、入力された情報の構造体。
type Input struct {
	ArticleID string `json:"article_id"`
	CommentID string `json:"comment_id"`
	UserName  string `json:"user_name"`
	Comment   string `json:"comment"`
}

var isLocal bool

func main() {
	t := flag.Bool("local", false, "ローカル実行か否か")
	articleID := flag.String("article-id", "", "ローカル実行用の記事ID")
	commentID := flag.String("comment-id", "", "ローカル実行用のコメントID")
	userName := flag.String("user-name", "", "ローカル実行用の表示名")
	comment := flag.String("comment", "", "ローカル実行用のコメント")
	flag.Parse()

	var err error
	isLocal, err = isLocalExec(t, articleID, commentID, userName, comment)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if isLocal {
		fmt.Println("local")
		localController(articleID, commentID, userName, comment)
		os.Exit(0)
	}

	fmt.Println("production")
	lambda.Start(controller)
}

// controller は、AWS Lambda 上での実行処理を行う
func controller(ctx context.Context, sqsEvent events.SQSEvent) error {
	inputs, err := inputs(sqsEvent)
	if err != nil {
		return err
	}

	for _, input := range inputs {
		if err := useCase(input); err != nil {
			return err
		}
	}

	return nil
}

// isLocalExec はローカル環境の実行であるかを判定する
func isLocalExec(t *bool, articleID *string, commentID *string, userName *string, comment *string) (bool, error) {
	if !*t {
		return false, nil
	}

	if *articleID == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だが記事ID指定が無いので処理不可能")
	}

	if *commentID == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だがコメントID指定が無いので処理不可能")
	}

	if *userName == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だが表示名指定が無いので処理不可能")
	}

	if *comment == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だがコメント指定が無いので処理不可能")
	}

	return true, nil
}

// localController はローカル環境での実行処理を行う
func localController(articleID *string, commentID *string, userName *string, comment *string) {
	input := Input{
		ArticleID: *articleID,
		CommentID: *commentID,
		UserName:  *userName,
		Comment:   *comment,
	}
	if err := useCase(input); err != nil {
		fmt.Println(err.Error())
	}
}

// useCase はアプリケーションのIFに依存しないメインの処理を行う
func useCase(input Input) error {
	return send(input)
}

// articleID は、SQSのイベント情報から記事IDを取得する
func inputs(sqsEvent events.SQSEvent) ([]Input, error) {
	if len(sqsEvent.Records) == 0 {
		return nil, errors.New("request content does not exist")
	}

	var inputs []Input
	for _, record := range sqsEvent.Records {
		b := []byte(record.Body)
		var input Input
		if err := json.Unmarshal(b, &input); err != nil {
			return nil, err
		}

		inputs = append(inputs, input)
	}

	return inputs, nil
}
