// Command api is the RinoFinance HTTP server entrypoint: it loads
// configuration, connects to MongoDB, ensures indexes exist, wires every
// hexagonal layer (repositories -> use cases -> handlers) and serves the
// REST API.
package main

import (
	"context"

	"github.com/google/uuid"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	rest "rinofinance-api/internal/adapters/http"
	"rinofinance-api/internal/adapters/http/handler"
	"rinofinance-api/internal/adapters/mongodb"
	appaccount "rinofinance-api/internal/application/account"
	appauth "rinofinance-api/internal/application/auth"
	appcard "rinofinance-api/internal/application/card"
	appcategory "rinofinance-api/internal/application/category"
	appdashboard "rinofinance-api/internal/application/dashboard"
	appexpense "rinofinance-api/internal/application/expense"
	appincome "rinofinance-api/internal/application/income"
	appinvestment "rinofinance-api/internal/application/investment"
	appprofile "rinofinance-api/internal/application/profile"
	appwishlist "rinofinance-api/internal/application/wishlist"
	"rinofinance-api/internal/config"
	pkgauth "rinofinance-api/internal/pkg/auth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar configuração: %v", err)
	}

	client, db, err := mongodb.Connect(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("erro ao conectar ao MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("erro ao desconectar do MongoDB: %v", err)
		}
	}()

	indexCtx, cancelIndexes := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelIndexes()
	if err := mongodb.EnsureIndexes(indexCtx, db); err != nil {
		log.Fatalf("erro ao garantir índices: %v", err)
	}
	log.Println("índices do MongoDB garantidos com sucesso")

	// Repositories (secondary/driven adapters).
	userRepo := mongodb.NewUserRepository(db)
	incomeRepo := mongodb.NewIncomeRepository(db)
	expenseRepo := mongodb.NewExpenseRepository(db)
	cardRepo := mongodb.NewCardRepository(db)
	installmentRepo := mongodb.NewInstallmentPurchaseRepository(db)
	subscriptionRepo := mongodb.NewSubscriptionRepository(db)
	investmentRepo := mongodb.NewInvestmentRepository(db)
	categoryRepo := mongodb.NewCategoryRepository(db)
	accountRepo := mongodb.NewAccountRepository(db)
	accountPurchaseRepo := mongodb.NewAccountPurchaseRepository(db)
	monthlyStatusRepo := mongodb.NewMonthlyStatusRepository(db)
	wishlistSectionRepo := mongodb.NewWishlistSectionRepository(db)
	wishlistItemRepo := mongodb.NewWishlistItemRepository(db)

	// Technical adapters used by the auth use cases.
	hasher := pkgauth.BcryptHasher{}
	tokens := pkgauth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)

	// Use cases (application layer).
	registerUser := appauth.NewRegisterUserUseCase(userRepo, hasher, cfg.RegistrationCode)
	loginUser := appauth.NewLoginUserUseCase(userRepo, hasher, tokens)

	updateProfile := appprofile.NewUpdateProfileUseCase(userRepo)
	changeEmail := appprofile.NewChangeEmailUseCase(userRepo, hasher)
	changePassword := appprofile.NewChangePasswordUseCase(userRepo, hasher)
	deleteAccount := appprofile.NewDeleteAccountUseCase(
		userRepo, hasher, incomeRepo, expenseRepo, cardRepo, installmentRepo, subscriptionRepo, investmentRepo,
	)
	shareData := appprofile.NewShareDataUseCase(userRepo, hasher)
	stopSharing := appprofile.NewStopSharingUseCase(userRepo)

	// Resolves the effective data owner for a user (self, or a shared account).
	resolveDataOwner := func(id uuid.UUID) uuid.UUID {
		if u, err := userRepo.FindByID(context.Background(), id); err == nil && u.DataOwnerID != nil {
			return *u.DataOwnerID
		}
		return id
	}

	accountBalanceResolver := appincome.NewAccountBalanceResolver(accountRepo)
	createIncome := appincome.NewCreateIncomeUseCase(incomeRepo)
	createAccountLinkedIncome := appincome.NewCreateAccountLinkedIncomeUseCase(incomeRepo, accountRepo)
	updateIncome := appincome.NewUpdateIncomeUseCase(incomeRepo)
	toggleIncome := appincome.NewToggleIncomeUseCase(incomeRepo)
	toggleIncomeReceived := appincome.NewToggleReceivedUseCase(incomeRepo, monthlyStatusRepo)
	deleteIncome := appincome.NewDeleteIncomeUseCase(incomeRepo)
	listIncomes := appincome.NewListIncomesUseCase(incomeRepo, accountBalanceResolver)
	reorderIncomes := appincome.NewReorderIncomesUseCase(incomeRepo)

	cardAmountResolver := appexpense.NewCardAmountResolver(installmentRepo, subscriptionRepo)
	accountLinkResolver := appexpense.NewAccountLinkResolver(accountPurchaseRepo)
	createExpense := appexpense.NewCreateExpenseUseCase(expenseRepo)
	createCardLinkedExpense := appexpense.NewCreateCardLinkedExpenseUseCase(expenseRepo, cardRepo)
	createAccountLinkedExpense := appexpense.NewCreateAccountLinkedExpenseUseCase(expenseRepo, accountRepo)
	updateExpense := appexpense.NewUpdateExpenseUseCase(expenseRepo)
	toggleExpense := appexpense.NewToggleExpenseUseCase(expenseRepo)
	togglePaidExpense := appexpense.NewTogglePaidUseCase(expenseRepo, monthlyStatusRepo)
	deleteExpense := appexpense.NewDeleteExpenseUseCase(expenseRepo)
	listExpenses := appexpense.NewListExpensesUseCase(expenseRepo, cardAmountResolver, accountLinkResolver)
	reorderExpenses := appexpense.NewReorderExpensesUseCase(expenseRepo)

	createCard := appcard.NewCreateCardUseCase(cardRepo)
	updateCard := appcard.NewUpdateCardUseCase(cardRepo)
	deleteCard := appcard.NewDeleteCardUseCase(cardRepo, installmentRepo, subscriptionRepo, expenseRepo)
	listCards := appcard.NewListCardsUseCase(cardRepo, installmentRepo, subscriptionRepo)
	createInstallmentPurchase := appcard.NewCreateInstallmentPurchaseUseCase(cardRepo, installmentRepo)
	updateInstallmentPurchase := appcard.NewUpdateInstallmentPurchaseUseCase(cardRepo, installmentRepo)
	toggleInstallmentFlag := appcard.NewToggleInstallmentPurchaseFlagUseCase(cardRepo, installmentRepo)
	toggleInstallmentOwed := appcard.NewToggleInstallmentPurchaseOwedUseCase(cardRepo, installmentRepo)
	deleteInstallmentPurchase := appcard.NewDeleteInstallmentPurchaseUseCase(cardRepo, installmentRepo)
	createSubscription := appcard.NewCreateSubscriptionUseCase(cardRepo, subscriptionRepo)
	updateSubscription := appcard.NewUpdateSubscriptionUseCase(cardRepo, subscriptionRepo)
	deleteSubscription := appcard.NewDeleteSubscriptionUseCase(cardRepo, subscriptionRepo)
	importCardItems := appcard.NewImportCardItemsUseCase(cardRepo, installmentRepo, subscriptionRepo)
	clearCardItems := appcard.NewClearCardItemsUseCase(cardRepo, installmentRepo, subscriptionRepo)
	reorderCards := appcard.NewReorderCardsUseCase(cardRepo)
	reorderInstallmentPurchases := appcard.NewReorderInstallmentPurchasesUseCase(cardRepo, installmentRepo)
	reorderSubscriptions := appcard.NewReorderSubscriptionsUseCase(cardRepo, subscriptionRepo)

	createCategory := appcategory.NewCreateCategoryUseCase(categoryRepo)
	updateCategory := appcategory.NewUpdateCategoryUseCase(categoryRepo)
	deleteCategory := appcategory.NewDeleteCategoryUseCase(categoryRepo)
	listCategories := appcategory.NewListCategoriesUseCase(categoryRepo)
	reorderCategories := appcategory.NewReorderCategoriesUseCase(categoryRepo)

	createAccount := appaccount.NewCreateAccountUseCase(accountRepo)
	updateAccount := appaccount.NewUpdateAccountUseCase(accountRepo)
	deleteAccount2 := appaccount.NewDeleteAccountUseCase(accountRepo, accountPurchaseRepo)
	listAccounts := appaccount.NewListAccountsUseCase(accountRepo, accountPurchaseRepo)
	reorderAccounts := appaccount.NewReorderAccountsUseCase(accountRepo)
	createAccountPurchase := appaccount.NewCreateAccountPurchaseUseCase(accountRepo, accountPurchaseRepo)
	updateAccountPurchase := appaccount.NewUpdateAccountPurchaseUseCase(accountRepo, accountPurchaseRepo)
	deleteAccountPurchase := appaccount.NewDeleteAccountPurchaseUseCase(accountRepo, accountPurchaseRepo)

	createAsset := appinvestment.NewCreateAssetUseCase(investmentRepo)
	updateAsset := appinvestment.NewUpdateAssetUseCase(investmentRepo)
	toggleAsset := appinvestment.NewToggleAssetUseCase(investmentRepo)
	deleteAsset := appinvestment.NewDeleteAssetUseCase(investmentRepo)
	listAssets := appinvestment.NewListAssetsUseCase(investmentRepo)

	wishlistService := appwishlist.NewService(wishlistSectionRepo, wishlistItemRepo)

	getMonthlySummary := appdashboard.NewGetMonthlySummaryUseCase(incomeRepo, expenseRepo, cardAmountResolver, accountLinkResolver, accountBalanceResolver, monthlyStatusRepo)

	// Handlers (primary/driving adapter).
	handlers := rest.Handlers{
		Auth:    handler.NewAuthHandler(registerUser, loginUser),
		Account: handler.NewAccountHandler(updateProfile, changeEmail, changePassword, deleteAccount, shareData, stopSharing),
		Income: handler.NewIncomeHandler(
			createIncome, createAccountLinkedIncome, updateIncome, toggleIncome, toggleIncomeReceived, deleteIncome,
			listIncomes, reorderIncomes,
		),
		Expense: handler.NewExpenseHandler(
			createExpense, createCardLinkedExpense, createAccountLinkedExpense, updateExpense, toggleExpense, togglePaidExpense,
			deleteExpense, listExpenses, reorderExpenses,
		),
		Card: handler.NewCardHandler(
			createCard, updateCard, deleteCard, listCards,
			createInstallmentPurchase, updateInstallmentPurchase, toggleInstallmentFlag, toggleInstallmentOwed, deleteInstallmentPurchase,
			createSubscription, updateSubscription, deleteSubscription, importCardItems, clearCardItems,
			reorderCards, reorderInstallmentPurchases, reorderSubscriptions,
		),
		Investment: handler.NewInvestmentHandler(createAsset, updateAsset, toggleAsset, deleteAsset, listAssets),
		Category:   handler.NewCategoryHandler(createCategory, updateCategory, deleteCategory, listCategories, reorderCategories),
		Wishlist:   handler.NewWishlistHandler(wishlistService),
		Wallet: handler.NewWalletHandler(
			createAccount, updateAccount, deleteAccount2, listAccounts, reorderAccounts,
			createAccountPurchase, updateAccountPurchase, deleteAccountPurchase,
		),
		Dashboard: handler.NewDashboardHandler(getMonthlySummary),
	}

	router := rest.NewRouter(handlers, tokens, resolveDataOwner, cfg.AllowedOrigins)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("rinofinance-api listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro no servidor HTTP: %v", err)
		}
	}()

	// Graceful shutdown: Railway sends SIGTERM before killing the
	// container on redeploy, so in-flight requests get a chance to finish.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("erro ao encerrar servidor graciosamente: %v", err)
	}
}
