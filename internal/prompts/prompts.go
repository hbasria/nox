// Package prompts holds nox's system prompts. Kept as a plain embedded text
// file so it stays easy to swap out later (e.g. merging with .crux agent
// prompts) without touching Go code.
package prompts

import _ "embed"

//go:embed system.txt
var system string

// System returns the base assistant system prompt.
func System() string {
	return system
}

// CommandGen appends task-specific instructions for turning a natural
// language request into a single runnable shell command.
func CommandGen() string {
	return system + `
Görev: Kullanıcının doğal dilde yazdığı isteği tek bir POSIX uyumlu shell komutuna çevir.
- SADECE komutu döndür. Açıklama, markdown code fence (` + "```" + `), veya ek metin ekleme.
- Çok adımlı işlemleri "&&" ile tek satırda birleştir.
- Emin değilsen en makul/en yaygın yorumu seç.`
}

// CommitMsg appends instructions for generating a commit message from a
// staged git diff.
func CommitMsg() string {
	return system + `
Görev: Verilen "git diff --staged" çıktısına göre kısa ve öz bir commit mesajı üret.
- SADECE commit mesajını döndür. Tırnak, açıklama, markdown ekleme.
- İlk satır 50-72 karakter civarı özet olsun, gerekiyorsa boş satır sonrası kısa açıklama ekle.
- Türkçe veya İngilizce fark etmez, diff'in dilini/tarzını yansıt; emin değilsen İngilizce yaz.`
}
