package usecase

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/Suhaibshah22/pipeweaver/internal/domain/entity"

	"gopkg.in/yaml.v2"
)

const TEMPLATE_BASE_PATH = "internal/usecase/templates/dag_template.py.tmpl"

type GenerateAirFlowDAGUsecase interface {
	Execute(ctx context.Context, pipelineFileContent []byte, filePath string) ([]byte, error)
}

type generateAirFlowDAGUsecase struct {
	Log *slog.Logger
}

func NewGenerateAirFlowDAGUsecase(
	logger *slog.Logger,
) GenerateAirFlowDAGUsecase {
	return &generateAirFlowDAGUsecase{
		Log: logger,
	}
}

type DAGTemplateData struct {
	PipelineName        string
	PipelineDescription string
	ScheduleInterval    string
	TaskName            string

	// Postgres
	PostgresHost     string
	PostgresDatabase string
	PostgresTable    string

	// Snowflake
	SnowflakeTable string
}

func (uc *generateAirFlowDAGUsecase) Execute(ctx context.Context, pipelineFileContent []byte, filePath string) ([]byte, error) {
	// 1. Parse the pipeline YAML
	upd, err := parseUPD(pipelineFileContent)
	if err != nil {
		uc.Log.Error("ParseUPD error", "error", err)
		return nil, fmt.Errorf("ParseUPD error: %w", err)
	}

	// 2. Prepare DAG template data
	dagData := DAGTemplateData{
		PipelineName:        upd.Pipeline.Name,
		PipelineDescription: upd.Pipeline.Description,
		ScheduleInterval:    getScheduleInterval(upd.Pipeline.Schedule),
		TaskName:            generateTaskName(upd.Pipeline.Steps),

		PostgresHost:     getDataRef(upd.Pipeline.Steps, "Postgres").Host,
		PostgresDatabase: getDataRef(upd.Pipeline.Steps, "Postgres").Database,
		PostgresTable:    getDataRef(upd.Pipeline.Steps, "Postgres").TableName,

		SnowflakeTable: getDataRef(upd.Pipeline.Steps, "Snowflake").TableName,
	}

	// 3. Determine Template Path based on version
	templatePath := TEMPLATE_BASE_PATH + "." + upd.Pipeline.Version
	uc.Log.Info("Pipeline template path", "info", templatePath)

	// 4. Generate DAG content
	return GenerateAirflowDAG(uc, dagData, templatePath)
}

func parseUPD(yamlData []byte) (*entity.UnifiedPipelineDefinition, error) {
	var upd entity.UnifiedPipelineDefinition
	err := yaml.Unmarshal(yamlData, &upd)
	if err != nil {
		return nil, err
	}
	return &upd, nil
}

func getScheduleInterval(schedule *entity.Schedule) string {
	if schedule == nil || schedule.Expression == "" {
		return "None"
	}
	return fmt.Sprintf(`"%s"`, schedule.Expression)
}

func generateTaskName(steps []entity.Step) string {
	if len(steps) == 0 {
		return "default_task"
	}
	return steps[0].Name
}

func getDataRef(steps []entity.Step, dataType string) entity.DataRef {
	for _, step := range steps {
		for _, input := range step.Inputs {
			if strings.Contains(strings.ToLower(input.Type), strings.ToLower(dataType)) {
				return input
			}
		}
		for _, output := range step.Outputs {
			if strings.Contains(strings.ToLower(output.Type), strings.ToLower(dataType)) {
				return output
			}
		}
	}
	return entity.DataRef{}
}

func GenerateAirflowDAG(uc *generateAirFlowDAGUsecase, data DAGTemplateData, templatePath string) ([]byte, error) {
	// 1. Read the template file
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("read template error: %w", err)
	}

	// 2. Parse the template
	tmpl, err := template.New("dag").Parse(string(tmplBytes))
	if err != nil {
		return nil, fmt.Errorf("parse template error: %w", err)
	}

	// 3. Execute the template
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return nil, fmt.Errorf("execute template error: %w", err)
	}

	uc.Log.Info("DAG generated successfully", "pipelineName", data.PipelineName)
	uc.Log.Debug("DAG content", "content", rendered.String())

	return rendered.Bytes(), nil
}
