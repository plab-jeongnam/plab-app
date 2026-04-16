package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/plab/plab-app/internal/model"
)

func IsUserAborted(err error) bool {
	return errors.Is(err, huh.ErrUserAborted)
}

var validNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func RunCreateForm() (*model.Project, error) {
	var rawName string
	var usePlabData bool
	var researchersOnly bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("어떤 프로젝트를 만드실 거에요?").
				Description("프로젝트 이름을 입력해 주세요 (예: landing, admin)").
				Placeholder("landing").
				Value(&rawName).
				Validate(func(s string) error {
					name := normalizeName(s)
					if name == "" {
						return fmt.Errorf("프로젝트 이름을 입력해 주세요")
					}
					if !validNamePattern.MatchString(name) {
						return fmt.Errorf("영문 소문자, 숫자, 하이픈(-)만 사용할 수 있어요")
					}
					fullName := name
					if !strings.HasPrefix(fullName, "plab-") {
						fullName = "plab-" + fullName
					}
					cwd, _ := os.Getwd()
					if _, err := os.Stat(filepath.Join(cwd, fullName)); err == nil {
						return fmt.Errorf("%s 폴더가 이미 있어요. 다른 이름을 입력해 주세요", fullName)
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("플랩의 데이터를 사용하실 거에요?").
				Description("예를 선택하면 플랩 API 연동 코드가 포함돼요").
				Affirmative("예").
				Negative("아니오").
				Value(&usePlabData),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("리서처들만 사용해야 하는 서비스에요?").
				Description("예를 선택하면 Google 로그인이 기본으로 설정돼요").
				Affirmative("예").
				Negative("아니오").
				Value(&researchersOnly),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	name := normalizeName(rawName)
	if !strings.HasPrefix(name, "plab-") {
		name = "plab-" + name
	}

	return &model.Project{
		Name:            name,
		DisplayName:     rawName,
		UsePlabData:     usePlabData,
		ResearchersOnly: researchersOnly,
	}, nil
}

func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.TrimPrefix(s, "plab-")
	return s
}
