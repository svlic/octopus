package op

import (
	"context"
	"fmt"
	"strings"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/utils/cache"
)

var llmModelCache = cache.New[string, model.LLMPrice](16)

// normalizeLLMName 统一模型名称键：去空白并小写，与 Create 写入规则一致。
func normalizeLLMName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func LLMList(ctx context.Context) ([]model.LLMInfo, error) {
	models := make([]model.LLMInfo, 0, llmModelCache.Len())
	for m, cost := range llmModelCache.GetAll() {
		models = append(models, model.LLMInfo{
			Name:     m,
			LLMPrice: cost,
		})
	}
	return models, nil
}

func LLMUpdate(m model.LLMInfo, ctx context.Context) error {
	m.Name = normalizeLLMName(m.Name)
	_, ok := llmModelCache.Get(m.Name)
	if !ok {
		return fmt.Errorf("model not found")
	}
	if err := db.GetDB().WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	llmModelCache.Set(m.Name, m.LLMPrice)
	return nil
}

func LLMDelete(modelName string, ctx context.Context) error {
	modelName = normalizeLLMName(modelName)
	_, ok := llmModelCache.Get(modelName)
	if !ok {
		return fmt.Errorf("model not found")
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.LLMInfo{Name: modelName}).Error; err != nil {
		return err
	}
	llmModelCache.Del(modelName)
	return nil
}
func LLMBatchDelete(modelNames []string, ctx context.Context) error {
	if len(modelNames) == 0 {
		return nil
	}
	names := make([]string, 0, len(modelNames))
	for _, n := range modelNames {
		n = normalizeLLMName(n)
		if n == "" {
			continue
		}
		names = append(names, n)
	}
	if len(names) == 0 {
		return nil
	}
	if err := db.GetDB().WithContext(ctx).Where("name IN ?", names).Delete(&model.LLMInfo{}).Error; err != nil {
		return err
	}
	llmModelCache.Del(names...)
	return nil
}
func LLMCreate(m model.LLMInfo, ctx context.Context) error {
	m.Name = normalizeLLMName(m.Name)
	_, ok := llmModelCache.Get(m.Name)
	if ok {
		return fmt.Errorf("model already exists")
	}
	if err := db.GetDB().WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	llmModelCache.Set(m.Name, m.LLMPrice)
	return nil
}
func LLMBatchCreate(llmInfos []model.LLMInfo, ctx context.Context) error {
	if len(llmInfos) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(llmInfos))
	newLLMInfos := make([]model.LLMInfo, 0, len(llmInfos))
	for _, llmInfo := range llmInfos {
		llmInfo.Name = normalizeLLMName(llmInfo.Name)
		if llmInfo.Name == "" {
			continue
		}
		if _, ok := seen[llmInfo.Name]; ok {
			continue
		}
		if _, ok := llmModelCache.Get(llmInfo.Name); ok {
			continue
		}
		seen[llmInfo.Name] = struct{}{}
		newLLMInfos = append(newLLMInfos, llmInfo)
	}
	if len(newLLMInfos) == 0 {
		return nil
	}
	if err := db.GetDB().WithContext(ctx).Create(&newLLMInfos).Error; err != nil {
		return err
	}
	for _, llmInfo := range newLLMInfos {
		llmModelCache.Set(llmInfo.Name, llmInfo.LLMPrice)
	}
	return nil
}
func LLMGet(name string) (model.LLMPrice, error) {
	price, ok := llmModelCache.Get(normalizeLLMName(name))
	if !ok {
		return model.LLMPrice{}, fmt.Errorf("model not found")
	}
	return price, nil
}

func llmRefreshCache(ctx context.Context) error {
	models := []model.LLMInfo{}
	if err := db.GetDB().WithContext(ctx).Find(&models).Error; err != nil {
		return err
	}
	for _, m := range models {
		llmModelCache.Set(normalizeLLMName(m.Name), m.LLMPrice)
	}
	return nil
}
