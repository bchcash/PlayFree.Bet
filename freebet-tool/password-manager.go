package main

import (
        "bufio"
        "context"
        "database/sql"
        "errors"
        "fmt"
        "log"
        "os"
        "os/user"
        "path/filepath"
        "strings"
        "time"

        _ "github.com/lib/pq" // Более простая библиотека для PostgreSQL
        "golang.org/x/crypto/bcrypt"
)

const (
        configFileName = ".user_passwords_backup.conf"
        saltRounds     = 10
)

type PasswordManager struct {
        db     *sql.DB
        config map[string]string
}

func NewPasswordManager() (*PasswordManager, error) {
        // Получаем домашнюю директорию
        usr, err := user.Current()
        if err != nil {
                return nil, fmt.Errorf("не удалось получить домашнюю директорию: %v", err)
        }

        configPath := filepath.Join(usr.HomeDir, configFileName)

        // Читаем конфигурационный файл
        config := make(map[string]string)
        if _, err := os.Stat(configPath); err == nil {
                file, err := os.Open(configPath)
                if err != nil {
                        return nil, fmt.Errorf("не удалось открыть файл конфигурации: %v", err)
                }
                defer file.Close()

                scanner := bufio.NewScanner(file)
                for scanner.Scan() {
                        line := scanner.Text()
                        if idx := strings.Index(line, "="); idx != -1 {
                                key := strings.TrimSpace(line[:idx])
                                value := strings.TrimSpace(line[idx+1:])
                                if key != "" && value != "" {
                                        config[key] = value
                                }
                        }
                }
                if err := scanner.Err(); err != nil {
                        return nil, fmt.Errorf("ошибка чтения файла конфигурации: %v", err)
                }
        }

        // Подключаемся к базе данных
        connStr := "postgres://ludic:2BjnKE63!@localhost:5432/ludic_db?sslmode=disable"
        db, err := sql.Open("postgres", connStr)
        if err != nil {
                return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
        }

        // Проверяем подключение
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := db.PingContext(ctx); err != nil {
                return nil, fmt.Errorf("не удалось подключиться к базе данных: %v", err)
        }

        return &PasswordManager{
                db:     db,
                config: config,
        }, nil
}

func (pm *PasswordManager) Close() error {
        return pm.db.Close()
}

func (pm *PasswordManager) saveConfig() error {
        usr, err := user.Current()
        if err != nil {
                return err
        }

        configPath := filepath.Join(usr.HomeDir, configFileName)

        if len(pm.config) == 0 {
                // Удаляем файл если нет конфигурации
                if _, err := os.Stat(configPath); err == nil {
                        return os.Remove(configPath)
                }
                return nil
        }

        file, err := os.Create(configPath)
        if err != nil {
                return err
        }
        defer file.Close()

        writer := bufio.NewWriter(file)
        for key, value := range pm.config {
                _, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
                if err != nil {
                        return err
                }
        }
        return writer.Flush()
}

func (pm *PasswordManager) BackupPassword(username string) error {
        fmt.Printf("Резервное копирование пароля для пользователя: %s\n", username)

        var currentHash string
        err := pm.db.QueryRow(
                "SELECT password_hash FROM users WHERE nickname = $1",
                username,
        ).Scan(&currentHash)

        if err != nil {
                if errors.Is(err, sql.ErrNoRows) {
                        return fmt.Errorf("пользователь '%s' не найден", username)
                }
                return fmt.Errorf("ошибка при запросе к базе данных: %v", err)
        }

        pm.config[username] = currentHash
        if err := pm.saveConfig(); err != nil {
                return fmt.Errorf("ошибка сохранения конфигурации: %v", err)
        }

        fmt.Printf("✓ Пароль пользователя %s сохранен\n", username)
        return nil
}

func (pm *PasswordManager) ResetPassword(username, tempPassword string) error {
        fmt.Printf("Сброс пароля для пользователя: %s\n", username)

        // Проверяем существование пользователя
        var count int
        err := pm.db.QueryRow(
                "SELECT COUNT(*) FROM users WHERE nickname = $1",
                username,
        ).Scan(&count)

        if err != nil {
                return fmt.Errorf("ошибка при проверке пользователя: %v", err)
        }

        if count == 0 {
                return fmt.Errorf("пользователь '%s' не найден", username)
        }

        // Делаем резервную копию
        if err := pm.BackupPassword(username); err != nil {
                return err
        }

        // Генерируем новый хеш
        fmt.Println("Генерация нового хеша для временного пароля...")
        newHash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), saltRounds)
        if err != nil {
                return fmt.Errorf("ошибка генерации хеша: %v", err)
        }

        // Обновляем пароль в базе данных
        fmt.Println("Обновление пароля в базе данных...")
        _, err = pm.db.Exec(
                "UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE nickname = $2",
                string(newHash),
                username,
        )

        if err != nil {
                return fmt.Errorf("ошибка обновления пароля: %v", err)
        }

        fmt.Printf("✓ Пароль для пользователя %s сброшен на временный\n", username)
        fmt.Printf("Временный пароль: %s\n", tempPassword)
        fmt.Println("⚠️  Обязательно сообщите этот пароль пользователю!")
        return nil
}

func (pm *PasswordManager) RestorePassword(username string) error {
        fmt.Printf("Восстановление исходного пароля для пользователя: %s\n", username)

        originalHash, exists := pm.config[username]
        if !exists {
                return fmt.Errorf("резервная копия пароля для пользователя %s не найдена", username)
        }

        // Восстанавливаем пароль
        _, err := pm.db.Exec(
                "UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE nickname = $2",
                originalHash,
                username,
        )

        if err != nil {
                return fmt.Errorf("ошибка восстановления пароля: %v", err)
        }

        // Удаляем из конфигурации
        delete(pm.config, username)
        if err := pm.saveConfig(); err != nil {
                return fmt.Errorf("ошибка сохранения конфигурации: %v", err)
        }

        fmt.Printf("✓ Исходный пароль для пользователя %s восстановлен\n", username)
        return nil
}

func (pm *PasswordManager) ListBackups() error {
        fmt.Println("Сохраненные резервные копии паролей:")

        if len(pm.config) == 0 {
                fmt.Println("Нет сохраненных резервных копий паролей")
                return nil
        }

        fmt.Println("=========================================")
        fmt.Println("Пользователь          | Время бэкапа")
        fmt.Println("-----------------------------------------")

        for username := range pm.config {
                var updatedAt time.Time
                err := pm.db.QueryRow(
                        "SELECT updated_at FROM users WHERE nickname = $1",
                        username,
                ).Scan(&updatedAt)

                if err != nil {
                        if errors.Is(err, sql.ErrNoRows) {
                                fmt.Printf("%-20s | Пользователь не найден в БД\n", username)
                        } else {
                                fmt.Printf("%-20s | Ошибка получения данных\n", username)
                        }
                } else {
                        fmt.Printf("%-20s | %s\n", username, updatedAt.Format("2006-01-02 15:04:05"))
                }
        }

        fmt.Println("=========================================")
        return nil
}

func (pm *PasswordManager) CheckUserStatus(username string) error {
        fmt.Printf("Проверка статуса пользователя: %s\n", username)

        type UserInfo struct {
                Nickname       string
                Email          string
                Money          float64
                CreatedAt      time.Time
                UpdatedAt      time.Time
                PasswordStatus string
        }

        var info UserInfo
        err := pm.db.QueryRow(`
                SELECT
                        nickname,
                        email,
                        money,
                        created_at,
                        updated_at,
                        CASE
                                WHEN password_hash LIKE '$2a$10$%' THEN 'Пароль сброшен (временный)'
                                ELSE 'Оригинальный пароль'
                        END as password_status
                FROM users
                WHERE nickname = $1`,
                username,
        ).Scan(&info.Nickname, &info.Email, &info.Money, &info.CreatedAt, &info.UpdatedAt, &info.PasswordStatus)

        if err != nil {
                if errors.Is(err, sql.ErrNoRows) {
                        return fmt.Errorf("пользователь не найден")
                }
                return fmt.Errorf("ошибка получения данных: %v", err)
        }

        fmt.Println("=========================================")
        fmt.Printf("Никнейм:      %s\n", info.Nickname)
        fmt.Printf("Email:        %s\n", info.Email)
        fmt.Printf("Баланс:       %.2f\n", info.Money)
        fmt.Printf("Создан:       %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("Обновлен:     %s\n", info.UpdatedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("Статус пароля: %s\n", info.PasswordStatus)
        fmt.Println("=========================================")
        return nil
}

func main() {
        if len(os.Args) < 2 {
                printUsage()
                os.Exit(1)
        }

        manager, err := NewPasswordManager()
        if err != nil {
                log.Fatalf("Ошибка инициализации: %v", err)
        }
        defer manager.Close()

        command := os.Args[1]

        switch command {
        case "reset":
                if len(os.Args) < 3 {
                        fmt.Println("Использование: reset <username> [temp-password]")
                        fmt.Println("Примеры:")
                        fmt.Println("  reset Alice")
                        fmt.Println("  reset Alice MyTempPass123")
                        os.Exit(1)
                }

                username := os.Args[2]
                tempPassword := "TempPass123!"

                // Парсим аргументы для временного пароля
                if len(os.Args) >= 4 {
                        // Проверяем, является ли третий аргумент паролем (не флагом)
                        if !strings.HasPrefix(os.Args[3], "-") {
                                tempPassword = os.Args[3]
                        } else {
                                // Если это флаг - проверяем формат -temp-password=value
                                for i := 3; i < len(os.Args); i++ {
                                        arg := os.Args[i]
                                        if strings.HasPrefix(arg, "-temp-password=") {
                                                tempPassword = strings.TrimPrefix(arg, "-temp-password=")
                                        } else if arg == "-p" && i+1 < len(os.Args) {
                                                tempPassword = os.Args[i+1]
                                        }
                                }
                        }
                }

                if err := manager.ResetPassword(username, tempPassword); err != nil {
                        log.Fatal(err)
                }

        case "restore":
                if len(os.Args) < 3 {
                        fmt.Println("Использование: restore <username>")
                        os.Exit(1)
                }
                username := os.Args[2]
                if err := manager.RestorePassword(username); err != nil {
                        log.Fatal(err)
                }

        case "list":
                if err := manager.ListBackups(); err != nil {
                        log.Fatal(err)
                }

        case "check":
                if len(os.Args) < 3 {
                        fmt.Println("Использование: check <username>")
                        os.Exit(1)
                }
                username := os.Args[2]
                if err := manager.CheckUserStatus(username); err != nil {
                        log.Fatal(err)
                }

        case "help":
                printUsage()

        default:
                fmt.Printf("Неизвестная команда: %s\n\n", command)
                printUsage()
                os.Exit(1)
        }
}

func printUsage() {
        fmt.Println("Менеджер паролей Freebet.Guru")
        fmt.Println("")
        fmt.Println("Использование:")
        fmt.Println("  reset <username> [temp-password]            - Сбросить пароль на временный")
        fmt.Println("  reset <username> [-temp-password=PASSWORD] - Сбросить пароль на временный")
        fmt.Println("  restore <username>                          - Восстановить оригинальный пароль")
        fmt.Println("  list                                        - Показать список резервных копий")
        fmt.Println("  check <username>                            - Проверить статус пользователя")
        fmt.Println("  help                                        - Показать эту справку")
        fmt.Println("")
        fmt.Println("Примеры:")
        fmt.Println("  ./password-manager reset Alice")
        fmt.Println("  ./password-manager reset Alice MyTempPass123")
        fmt.Println("  ./password-manager reset Alice -temp-password=MyTempPass123")
        fmt.Println("  ./password-manager restore Alice")
        fmt.Println("  ./password-manager list")
        fmt.Println("  ./password-manager check Alice")
}
