package subscription

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/invoice"
	"gorm.io/gorm"

	"github.com/hngprojects/telex_be/internal/models"
	"github.com/hngprojects/telex_be/pkg/middleware"
)

var (
	ErrNotOrgMember       = errors.New("user is not a member of this organisation")
	ErrSessionOrgMismatch = errors.New("stripe session does not belong to this organisation")
)

func DownloadInvoice(sessionID string, c *gin.Context, db *gorm.DB, rdb *redis.Client, orgID string) (gin.H, error) {

	userIdVal, err := middleware.GetUserClaims(c, db, "user_id")
	if err != nil {
		return gin.H{}, errors.New("user context required")
	}
	userID, ok := userIdVal.(string)
	if !ok || userID == "" {
		return gin.H{}, errors.New("invalid user context")
	}

	var org models.Organisation
	isMember, err := org.CheckUserIsMemberOfOrg(userID, orgID, db)
	if err != nil || !isMember {
		return gin.H{}, ErrNotOrgMember
	}

	req := models.CompleteSubscriptionRequest{
		StripeSessionID: sessionID,
	}

	sesh, err := session.Get(req.StripeSessionID, nil)
	if err != nil {
		return gin.H{}, err
	}

	sessionOrgID := sesh.Metadata["org_id"]
	if sessionOrgID != "" && sessionOrgID != orgID {
		return gin.H{}, ErrSessionOrgMismatch
	}

	if sesh.Subscription == nil || sesh.Subscription.ID == "" {
		return gin.H{}, errors.New("no subscription ID found for this session")
	}

	params := &stripe.InvoiceListParams{
		Subscription: stripe.String(sesh.Subscription.ID),
	}

	i := invoice.List(params)

	var invoiceItems []*stripe.Invoice

	for i.Next() {
		invoiceItems = append(invoiceItems, i.Invoice())
	}

	if len(invoiceItems) == 0 {
		return gin.H{}, errors.New("no invoices found for this subscription")
	}

	pdfURL := invoiceItems[0].InvoicePDF
	if pdfURL == "" {
		return gin.H{}, errors.New("PDF not available for this invoice")
	}

	url := gin.H{
		"invoice_url": pdfURL,
	}

	return url, nil
}
