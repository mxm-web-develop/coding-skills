package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreSkillsUseUserCommunicationContract(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	contractPath := filepath.Join(repositoryRoot, "skills", "orchestrate-ai-delivery", "references", "user-communication-contract.md")
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	contractText := string(contract)
	for _, required := range []string{"3MS+7WI", "三个阶段，共七项开发任务", "Translate before speaking", "Progress and long tasks"} {
		if !strings.Contains(contractText, required) {
			t.Fatalf("user communication contract is missing %q", required)
		}
	}

	for _, skill := range coreSkills {
		path := filepath.Join(repositoryRoot, "skills", skill, "SKILL.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "user-communication-contract.md") {
			t.Errorf("%s does not reference the user communication contract", skill)
		}

		metadataPath := filepath.Join(repositoryRoot, "skills", skill, "agents", "openai.yaml")
		metadata, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatal(err)
		}
		metadataText := string(metadata)
		if strings.Contains(metadataText, "default_prompt:") || strings.Contains(metadataText, "$"+skill) {
			t.Errorf("%s user-facing metadata exposes an internal skill invocation", skill)
		}
		for _, forbidden := range []string{"Run the ", "AI Flow workflow", "Work Item", "playbook"} {
			if strings.Contains(metadataText, forbidden) {
				t.Errorf("%s user-facing metadata contains internal wording %q", skill, forbidden)
			}
		}
	}
}

func TestPlatformEntriesRequireNaturalUserLanguage(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	for _, relativePath := range []string{
		"adapters/codex/AGENTS.block.md",
		"adapters/cursor/ai-flow.mdc",
		"adapters/claude/CLAUDE.block.md",
	} {
		content, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(relativePath)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "Keep AI Flow implementation vocabulary internal") {
			t.Errorf("%s does not enforce natural user-facing language", relativePath)
		}
	}
}

func TestConversationContinuityKeepsInterruptionsAndApprovalsSeparate(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	continuityPath := filepath.Join(repositoryRoot, "skills", "orchestrate-ai-delivery", "references", "conversation-continuity.md")
	continuity, err := os.ReadFile(continuityPath)
	if err != nil {
		t.Fatal(err)
	}
	continuityText := string(continuity)
	for _, required := range []string{
		"Status or explanation question", "Material addition or changed acceptance", "Unrelated side question",
		"Resume request or IDE switch", "A reply counts as approval only when it clearly answers the exact pending choice",
		"does not invalidate completed work", "what can be reused",
	} {
		if !strings.Contains(continuityText, required) {
			t.Errorf("conversation continuity contract is missing %q", required)
		}
	}
	orchestrator, err := os.ReadFile(filepath.Join(repositoryRoot, "skills", "orchestrate-ai-delivery", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(orchestrator), "references/conversation-continuity.md") {
		t.Fatal("orchestrator does not route interrupted conversations through the continuity contract")
	}
}
