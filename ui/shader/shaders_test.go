package shader

import "testing"

func TestLoadAll(t *testing.T) {
	for _, name := range Names() {
		shader, err := Load(name)
		if err != nil {
			t.Errorf("failed to load shader %q: %v", name, err)
			continue
		}
		shader.Deallocate()
	}
}

func TestDefaultExists(t *testing.T) {
	_, err := Load(Default)
	if err != nil {
		t.Errorf("failed to load default shader %q: %v", Default, err)

	}
}
