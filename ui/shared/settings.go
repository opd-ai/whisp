package shared

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/whisp/internal/core/config"
)

// SettingsDialog represents the settings configuration interface
// Provides tabbed interface for different setting categories
type SettingsDialog struct {
	dialog       *dialog.CustomDialog
	configMgr    *config.Manager
	parentWindow fyne.Window

	// UI bindings for real-time updates
	themeBinding    binding.String
	fontSizeBinding binding.String
	languageBinding binding.String
}

// NewSettingsDialog creates a new settings dialog
// Takes config manager and parent window for modal behavior
func NewSettingsDialog(configMgr *config.Manager, parentWindow fyne.Window) *SettingsDialog {
	sd := &SettingsDialog{
		configMgr:    configMgr,
		parentWindow: parentWindow,
	}

	// Create data bindings for real-time UI updates
	sd.themeBinding = binding.NewString()
	sd.fontSizeBinding = binding.NewString()
	sd.languageBinding = binding.NewString()

	// Load current values into bindings
	cfg := configMgr.GetConfig()
	sd.themeBinding.Set(cfg.UI.Theme)
	sd.fontSizeBinding.Set(cfg.UI.FontSize)
	sd.languageBinding.Set(cfg.UI.Language)

	return sd
}

// Show displays the settings dialog
// Creates modal dialog with save/cancel buttons
func (sd *SettingsDialog) Show() {
	content := sd.createContent()

	// Create dialog with save and cancel buttons
	sd.dialog = dialog.NewCustom("Settings", "Close", content, sd.parentWindow)
	sd.dialog.Resize(fyne.NewSize(600, 500))
	sd.dialog.Show()
}

// createContent creates the main content with tabs
// Organizes settings into logical groups for usability
func (sd *SettingsDialog) createContent() fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabWithText("General", sd.createGeneralTab()),
		container.NewTabWithText("Privacy", sd.createPrivacyTab()),
		container.NewTabWithText("Notifications", sd.createNotificationsTab()),
		container.NewTabWithText("Advanced", sd.createAdvancedTab()),
	)

	// Add save/apply buttons at the bottom
	saveBtn := widget.NewButton("Save", sd.saveSettings)
	saveBtn.Importance = widget.HighImportance

	applyBtn := widget.NewButton("Apply", sd.applySettings)
	resetBtn := widget.NewButton("Reset to Defaults", sd.resetToDefaults)

	buttonContainer := container.NewHBox(
		widget.NewLabel(""), // Spacer
		resetBtn,
		applyBtn,
		saveBtn,
	)

	return container.NewBorder(
		nil,             // top
		buttonContainer, // bottom
		nil,             // left
		nil,             // right
		tabs,            // center
	)
}

// createGeneralTab creates the general settings tab
// Includes UI, storage, and basic configuration options
func (sd *SettingsDialog) createGeneralTab() fyne.CanvasObject {
	cfg := sd.configMgr.GetConfig()

	// Theme selection
	themeSelect := widget.NewSelect(
		[]string{"system", "light", "dark", "amoled"},
		func(value string) {
			sd.themeBinding.Set(value)
		},
	)
	themeSelect.SetSelected(cfg.UI.Theme)

	// Font size selection
	fontSizeSelect := widget.NewSelect(
		[]string{"small", "medium", "large", "extra_large"},
		func(value string) {
			sd.fontSizeBinding.Set(value)
		},
	)
	fontSizeSelect.SetSelected(cfg.UI.FontSize)

	// Language selection (basic for now)
	languageSelect := widget.NewSelect(
		[]string{"en", "es", "fr", "de", "pt", "ru", "zh"},
		func(value string) {
			sd.languageBinding.Set(value)
		},
	)
	languageSelect.SetSelected(cfg.UI.Language)

	// Storage settings
	encryptionCheck := widget.NewCheck("Enable database encryption", nil)
	encryptionCheck.SetChecked(cfg.Storage.EnableEncryption)

	animationsCheck := widget.NewCheck("Enable animations", nil)
	animationsCheck.SetChecked(cfg.UI.EnableAnimations)

	soundCheck := widget.NewCheck("Enable sound effects", nil)
	soundCheck.SetChecked(cfg.UI.EnableSoundEffects)

	// File size limit
	maxFileSizeEntry := widget.NewEntry()
	maxFileSizeEntry.SetText(fmt.Sprintf("%.0f", float64(cfg.Storage.MaxFileSize)/(1024*1024*1024))) // Convert to GB

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Theme", themeSelect),
			widget.NewFormItem("Font Size", fontSizeSelect),
			widget.NewFormItem("Language", languageSelect),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Database Encryption", encryptionCheck),
			widget.NewFormItem("Animations", animationsCheck),
			widget.NewFormItem("Sound Effects", soundCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Max File Size (GB)", maxFileSizeEntry),
		},
	}

	// Store references for saving
	sd.storeFormReferences("general", map[string]interface{}{
		"theme":       themeSelect,
		"fontSize":    fontSizeSelect,
		"language":    languageSelect,
		"encryption":  encryptionCheck,
		"animations":  animationsCheck,
		"sound":       soundCheck,
		"maxFileSize": maxFileSizeEntry,
	})

	return container.NewScroll(form)
}

// createPrivacyTab creates the privacy settings tab
// Includes message history, typing indicators, and security options
func (sd *SettingsDialog) createPrivacyTab() fyne.CanvasObject {
	cfg := sd.configMgr.GetConfig()

	// Message privacy
	saveHistoryCheck := widget.NewCheck("Save message history", nil)
	saveHistoryCheck.SetChecked(cfg.Privacy.SaveMessageHistory)

	disappearingCheck := widget.NewCheck("Enable disappearing messages", nil)
	disappearingCheck.SetChecked(cfg.Privacy.EnableDisappearingMessages)

	// Typing indicators
	showTypingCheck := widget.NewCheck("Show typing indicators", nil)
	showTypingCheck.SetChecked(cfg.Privacy.ShowTypingIndicators)

	sendTypingCheck := widget.NewCheck("Send typing indicators", nil)
	sendTypingCheck.SetChecked(cfg.Privacy.SendTypingIndicators)

	// Read receipts
	showReceiptsCheck := widget.NewCheck("Show read receipts", nil)
	showReceiptsCheck.SetChecked(cfg.Privacy.ShowReadReceipts)

	sendReceiptsCheck := widget.NewCheck("Send read receipts", nil)
	sendReceiptsCheck.SetChecked(cfg.Privacy.SendReadReceipts)

	// File sharing
	autoAcceptCheck := widget.NewCheck("Auto-accept files from friends", nil)
	autoAcceptCheck.SetChecked(cfg.Privacy.AutoAcceptFiles)

	autoDownloadEntry := widget.NewEntry()
	autoDownloadEntry.SetText(fmt.Sprintf("%.0f", float64(cfg.Privacy.AutoDownloadLimit)/(1024*1024))) // Convert to MB

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Message History", saveHistoryCheck),
			widget.NewFormItem("Disappearing Messages", disappearingCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Show Typing Status", showTypingCheck),
			widget.NewFormItem("Send Typing Status", sendTypingCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Show Read Receipts", showReceiptsCheck),
			widget.NewFormItem("Send Read Receipts", sendReceiptsCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Auto-Accept Files", autoAcceptCheck),
			widget.NewFormItem("Auto-Download Limit (MB)", autoDownloadEntry),
		},
	}

	sd.storeFormReferences("privacy", map[string]interface{}{
		"saveHistory":  saveHistoryCheck,
		"disappearing": disappearingCheck,
		"showTyping":   showTypingCheck,
		"sendTyping":   sendTypingCheck,
		"showReceipts": showReceiptsCheck,
		"sendReceipts": sendReceiptsCheck,
		"autoAccept":   autoAcceptCheck,
		"autoDownload": autoDownloadEntry,
	})

	return container.NewScroll(form)
}

// createNotificationsTab creates the notifications settings tab
// Includes desktop and mobile notification preferences
func (sd *SettingsDialog) createNotificationsTab() fyne.CanvasObject {
	cfg := sd.configMgr.GetConfig()

	// Global notifications
	enabledCheck := widget.NewCheck("Enable notifications", nil)
	enabledCheck.SetChecked(cfg.Notifications.Enabled)

	// Desktop notifications
	desktopPreviewCheck := widget.NewCheck("Show message preview", nil)
	desktopPreviewCheck.SetChecked(cfg.Notifications.Desktop.ShowPreview)

	desktopSoundCheck := widget.NewCheck("Play notification sound", nil)
	desktopSoundCheck.SetChecked(cfg.Notifications.Desktop.PlaySound)

	desktopSenderCheck := widget.NewCheck("Show sender name", nil)
	desktopSenderCheck.SetChecked(cfg.Notifications.Desktop.ShowSender)

	// Mobile notifications
	mobileVibrateCheck := widget.NewCheck("Vibrate on message", nil)
	mobileVibrateCheck.SetChecked(cfg.Notifications.Mobile.Vibrate)

	mobileLockScreenCheck := widget.NewCheck("Show on lock screen", nil)
	mobileLockScreenCheck.SetChecked(cfg.Notifications.Mobile.ShowOnLockScreen)

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Enable Notifications", enabledCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Desktop: Show Preview", desktopPreviewCheck),
			widget.NewFormItem("Desktop: Play Sound", desktopSoundCheck),
			widget.NewFormItem("Desktop: Show Sender", desktopSenderCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Mobile: Vibrate", mobileVibrateCheck),
			widget.NewFormItem("Mobile: Lock Screen", mobileLockScreenCheck),
		},
	}

	sd.storeFormReferences("notifications", map[string]interface{}{
		"enabled":          enabledCheck,
		"desktopPreview":   desktopPreviewCheck,
		"desktopSound":     desktopSoundCheck,
		"desktopSender":    desktopSenderCheck,
		"mobileVibrate":    mobileVibrateCheck,
		"mobileLockScreen": mobileLockScreenCheck,
	})

	return container.NewScroll(form)
}

// createAdvancedTab creates the advanced settings tab
// Includes logging, debugging, and experimental features
func (sd *SettingsDialog) createAdvancedTab() fyne.CanvasObject {
	cfg := sd.configMgr.GetConfig()

	// Logging
	logLevelSelect := widget.NewSelect(
		[]string{"debug", "info", "warn", "error"},
		nil,
	)
	logLevelSelect.SetSelected(cfg.Advanced.LogLevel)

	logToFileCheck := widget.NewCheck("Log to file", nil)
	logToFileCheck.SetChecked(cfg.Advanced.LogToFile)

	debugModeCheck := widget.NewCheck("Enable debug mode", nil)
	debugModeCheck.SetChecked(cfg.Advanced.EnableDebugMode)

	// Performance
	maxDownloadsEntry := widget.NewEntry()
	maxDownloadsEntry.SetText(strconv.Itoa(cfg.Advanced.MaxConcurrentDownloads))

	maxUploadsEntry := widget.NewEntry()
	maxUploadsEntry.SetText(strconv.Itoa(cfg.Advanced.MaxConcurrentUploads))

	cacheSizeEntry := widget.NewEntry()
	cacheSizeEntry.SetText(strconv.Itoa(cfg.Advanced.MessageCacheSize))

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Log Level", logLevelSelect),
			widget.NewFormItem("Log to File", logToFileCheck),
			widget.NewFormItem("Debug Mode", debugModeCheck),
			widget.NewFormItem("", widget.NewSeparator()),
			widget.NewFormItem("Max Concurrent Downloads", maxDownloadsEntry),
			widget.NewFormItem("Max Concurrent Uploads", maxUploadsEntry),
			widget.NewFormItem("Message Cache Size", cacheSizeEntry),
		},
	}

	sd.storeFormReferences("advanced", map[string]interface{}{
		"logLevel":     logLevelSelect,
		"logToFile":    logToFileCheck,
		"debugMode":    debugModeCheck,
		"maxDownloads": maxDownloadsEntry,
		"maxUploads":   maxUploadsEntry,
		"cacheSize":    cacheSizeEntry,
	})

	return container.NewScroll(form)
}

// Form references storage for accessing values during save
var formReferences = make(map[string]map[string]interface{})

// storeFormReferences stores widget references for later access
func (sd *SettingsDialog) storeFormReferences(section string, refs map[string]interface{}) {
	formReferences[section] = refs
}

// saveSettings saves the current settings and closes the dialog
func (sd *SettingsDialog) saveSettings() {
	if err := sd.applySettings(); err != nil {
		dialog.ShowError(err, sd.parentWindow)
		return
	}
	sd.dialog.Hide()
}

// applySettings applies the current form values to configuration
func (sd *SettingsDialog) applySettings() error {
	cfg := sd.configMgr.GetConfig()

	// Apply general settings
	if general, ok := formReferences["general"]; ok {
		if theme, ok := general["theme"].(*widget.Select); ok {
			cfg.UI.Theme = theme.Selected
		}
		if fontSize, ok := general["fontSize"].(*widget.Select); ok {
			cfg.UI.FontSize = fontSize.Selected
		}
		if language, ok := general["language"].(*widget.Select); ok {
			cfg.UI.Language = language.Selected
		}
		if encryption, ok := general["encryption"].(*widget.Check); ok {
			cfg.Storage.EnableEncryption = encryption.Checked
		}
		if animations, ok := general["animations"].(*widget.Check); ok {
			cfg.UI.EnableAnimations = animations.Checked
		}
		if sound, ok := general["sound"].(*widget.Check); ok {
			cfg.UI.EnableSoundEffects = sound.Checked
		}
		if maxFileSize, ok := general["maxFileSize"].(*widget.Entry); ok {
			if size, err := strconv.ParseFloat(maxFileSize.Text, 64); err == nil {
				cfg.Storage.MaxFileSize = int64(size * 1024 * 1024 * 1024) // Convert GB to bytes
			}
		}
	}

	// Apply privacy settings
	if privacy, ok := formReferences["privacy"]; ok {
		if saveHistory, ok := privacy["saveHistory"].(*widget.Check); ok {
			cfg.Privacy.SaveMessageHistory = saveHistory.Checked
		}
		if disappearing, ok := privacy["disappearing"].(*widget.Check); ok {
			cfg.Privacy.EnableDisappearingMessages = disappearing.Checked
		}
		if showTyping, ok := privacy["showTyping"].(*widget.Check); ok {
			cfg.Privacy.ShowTypingIndicators = showTyping.Checked
		}
		if sendTyping, ok := privacy["sendTyping"].(*widget.Check); ok {
			cfg.Privacy.SendTypingIndicators = sendTyping.Checked
		}
		if showReceipts, ok := privacy["showReceipts"].(*widget.Check); ok {
			cfg.Privacy.ShowReadReceipts = showReceipts.Checked
		}
		if sendReceipts, ok := privacy["sendReceipts"].(*widget.Check); ok {
			cfg.Privacy.SendReadReceipts = sendReceipts.Checked
		}
		if autoAccept, ok := privacy["autoAccept"].(*widget.Check); ok {
			cfg.Privacy.AutoAcceptFiles = autoAccept.Checked
		}
		if autoDownload, ok := privacy["autoDownload"].(*widget.Entry); ok {
			if size, err := strconv.ParseFloat(autoDownload.Text, 64); err == nil {
				cfg.Privacy.AutoDownloadLimit = int64(size * 1024 * 1024) // Convert MB to bytes
			}
		}
	}

	// Apply notification settings
	if notifications, ok := formReferences["notifications"]; ok {
		if enabled, ok := notifications["enabled"].(*widget.Check); ok {
			cfg.Notifications.Enabled = enabled.Checked
		}
		if desktopPreview, ok := notifications["desktopPreview"].(*widget.Check); ok {
			cfg.Notifications.Desktop.ShowPreview = desktopPreview.Checked
		}
		if desktopSound, ok := notifications["desktopSound"].(*widget.Check); ok {
			cfg.Notifications.Desktop.PlaySound = desktopSound.Checked
		}
		if desktopSender, ok := notifications["desktopSender"].(*widget.Check); ok {
			cfg.Notifications.Desktop.ShowSender = desktopSender.Checked
		}
		if mobileVibrate, ok := notifications["mobileVibrate"].(*widget.Check); ok {
			cfg.Notifications.Mobile.Vibrate = mobileVibrate.Checked
		}
		if mobileLockScreen, ok := notifications["mobileLockScreen"].(*widget.Check); ok {
			cfg.Notifications.Mobile.ShowOnLockScreen = mobileLockScreen.Checked
		}
	}

	// Apply advanced settings
	if advanced, ok := formReferences["advanced"]; ok {
		if logLevel, ok := advanced["logLevel"].(*widget.Select); ok {
			cfg.Advanced.LogLevel = logLevel.Selected
		}
		if logToFile, ok := advanced["logToFile"].(*widget.Check); ok {
			cfg.Advanced.LogToFile = logToFile.Checked
		}
		if debugMode, ok := advanced["debugMode"].(*widget.Check); ok {
			cfg.Advanced.EnableDebugMode = debugMode.Checked
		}
		if maxDownloads, ok := advanced["maxDownloads"].(*widget.Entry); ok {
			if count, err := strconv.Atoi(maxDownloads.Text); err == nil {
				cfg.Advanced.MaxConcurrentDownloads = count
			}
		}
		if maxUploads, ok := advanced["maxUploads"].(*widget.Entry); ok {
			if count, err := strconv.Atoi(maxUploads.Text); err == nil {
				cfg.Advanced.MaxConcurrentUploads = count
			}
		}
		if cacheSize, ok := advanced["cacheSize"].(*widget.Entry); ok {
			if size, err := strconv.Atoi(cacheSize.Text); err == nil {
				cfg.Advanced.MessageCacheSize = size
			}
		}
	}

	// Save configuration
	return sd.configMgr.UpdateConfig(cfg)
}

// resetToDefaults resets all settings to default values
func (sd *SettingsDialog) resetToDefaults() {
	dialog.ShowConfirm(
		"Reset Settings",
		"Are you sure you want to reset all settings to their default values? This action cannot be undone.",
		func(confirmed bool) {
			if confirmed {
				// Create new config manager to get defaults
				if tempMgr, err := config.NewManager(""); err == nil {
					defaultCfg := tempMgr.GetConfig()
					if err := sd.configMgr.UpdateConfig(defaultCfg); err != nil {
						dialog.ShowError(fmt.Errorf("failed to reset settings: %w", err), sd.parentWindow)
						return
					}
					// Close and reopen dialog to refresh values
					sd.dialog.Hide()
					NewSettingsDialog(sd.configMgr, sd.parentWindow).Show()
				}
			}
		},
		sd.parentWindow,
	)
}
