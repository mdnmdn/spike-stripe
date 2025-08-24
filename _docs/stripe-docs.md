# Stripe Backend Integration Guide

## Table of Contents
- [Overview](#overview)
- [Authentication & API Keys](#authentication--api-keys)
- [Core Concepts](#core-concepts)
- [Customer Management](#customer-management)
- [Payment Methods](#payment-methods)
- [Payment Intents](#payment-intents)
- [Subscriptions & Billing](#subscriptions--billing)
- [Webhooks](#webhooks)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Code Examples](#code-examples)

## Overview

Stripe provides a comprehensive payment processing platform with RESTful APIs for backend integration. The API is organized around REST principles, accepts form-encoded requests, returns JSON-encoded responses, and uses standard HTTP response codes and authentication.

### Key Features
- Online payments with multiple integration approaches
- Subscription billing and invoicing
- International payment support
- Platform and marketplace payments
- Strong Customer Authentication (SCA) compliance
- Comprehensive webhook system

## Authentication & API Keys

### API Key Types
- **Test Keys**: Start with `sk_test_` - for development and testing
- **Live Keys**: Start with `sk_live_` - for production use
- **Account Keys**: Start with `acct_` - for connected accounts

### Authentication Methods

#### HTTP Basic Auth
```bash
curl https://api.stripe.com/v1/charges \
  -u sk_test_...:
```

#### Bearer Authentication
```bash
curl https://api.stripe.com/v1/charges \
  -H "Authorization: Bearer sk_test_..."
```

#### Library Configuration
```python
# Python
import stripe
stripe.api_key = "sk_test_..."
```

```javascript
// Node.js
const stripe = require('stripe')('sk_test_...');
```

### Connected Accounts
Use the `Stripe-Account` header to make API calls on behalf of connected accounts:
```bash
curl https://api.stripe.com/v1/charges \
  -H "Authorization: Bearer sk_test_..." \
  -H "Stripe-Account: acct_..."
```

### Security Requirements
- **All API requests must be made over HTTPS**
- Keep API keys secure and never expose them in client-side code
- Rotate keys regularly in production environments
- Use environment variables to store API keys

## Core Concepts

### Key Objects
- **Customer**: Represents a customer and stores payment methods
- **PaymentIntent**: Manages the lifecycle of a payment
- **PaymentMethod**: Represents a payment method (card, bank account, etc.)
- **Invoice**: Tracks billing amounts and payment status
- **Subscription**: Represents recurring purchases
- **Product**: Defines the service or good being sold
- **Price**: Defines billing details for products

### API Characteristics
- No bulk updates supported - each request works on one object
- Uses idempotency keys to prevent duplicate operations
- Functionality can vary across account versions
- Metadata can be attached to most objects for custom tracking

## Customer Management

### Creating Customers
```python
customer = stripe.Customer.create(
    email='customer@example.com',
    name='John Doe',
    metadata={'user_id': '123'}
)
```

### Customer Benefits
- Save and reuse payment methods
- Track multiple payments
- Manage subscriptions
- Store customer information securely

### Customer Sessions
Create CustomerSession objects for frontend integration:
```python
customer_session = stripe.CustomerSession.create(
    customer=customer.id,
    components={
        'payment_element': {
            'enabled': True,
            'features': {
                'payment_method_save': 'enabled',
                'payment_method_redisplay': 'enabled'
            }
        }
    }
)
```

## Payment Methods

### Supported Types
- Credit/Debit Cards
- Bank debits (ACH, SEPA, etc.)
- Bank redirects (iDEAL, Bancontact, etc.)
- Digital wallets (Apple Pay, Google Pay, etc.)
- Cryptocurrency options
- Buy now, pay later services

### Saving Payment Methods

#### Using SetupIntent (without initial payment)
```python
setup_intent = stripe.SetupIntent.create(
    customer=customer.id,
    payment_method_types=['card'],
    usage='off_session'
)
```

#### Automatic Saving During Payment
```python
payment_intent = stripe.PaymentIntent.create(
    amount=2000,
    currency='usd',
    customer=customer.id,
    setup_future_usage='off_session'
)
```

### Compliance Considerations
- Obtain explicit customer consent before saving payment methods
- Clearly communicate how payment methods will be used
- Comply with applicable laws and regulations (PCI DSS, GDPR, etc.)

## Payment Intents

### Overview
PaymentIntents manage complex payment flows with changing states throughout the transaction lifecycle. They provide:
- Automatic authentication management
- Prevention of double charging
- Strong Customer Authentication (SCA) support
- Idempotency key handling

### Creating Payment Intents
```python
payment_intent = stripe.PaymentIntent.create(
    amount=2000,  # Amount in cents
    currency='usd',
    customer=customer.id,
    metadata={'order_id': 'order_123'},
    statement_descriptor='ACME Store'
)
```

### Payment Intent Lifecycle
1. **Create** PaymentIntent when amount is known
2. **Pass** client secret to frontend
3. **Confirm** payment using Stripe.js
4. **Monitor** webhook events for status updates

### Status Values
- `requires_payment_method`: Needs payment method
- `requires_confirmation`: Ready to be confirmed
- `requires_action`: Needs additional authentication
- `processing`: Payment is being processed
- `succeeded`: Payment completed successfully
- `canceled`: Payment was canceled

### Best Practices
- Reuse the same PaymentIntent if payment process is interrupted
- Use idempotency keys to prevent duplicate PaymentIntents
- Store PaymentIntent ID for order tracking
- Handle additional authentication flows properly

## Subscriptions & Billing

### Subscription Lifecycle

#### 1. Creation
```python
subscription = stripe.Subscription.create(
    customer=customer.id,
    items=[{
        'price': 'price_1234',
    }],
    payment_behavior='default_incomplete',
    payment_settings={
        'payment_method_options': {
            'card': {
                'request_three_d_secure': 'any'
            }
        }
    }
)
```

#### 2. Status Management
- `trialing`: Free trial period
- `active`: Subscription in good standing
- `incomplete`: Payment pending or requires authentication
- `past_due`: Payment failed but retrying
- `canceled`: Subscription terminated
- `unpaid`: Latest invoice remains unresolved

### Billing Features
- **Smart Retries**: Automatic retry logic for failed payments
- **Proration**: Automatic calculations for mid-cycle changes
- **Trial Periods**: Free trial support
- **Discounts**: Coupon and discount management
- **Usage-based Billing**: Metered billing support

### Managing Subscriptions
```python
# Update subscription
stripe.Subscription.modify(
    subscription.id,
    items=[{
        'id': subscription['items']['data'][0]['id'],
        'price': 'new_price_id',
    }]
)

# Cancel subscription
stripe.Subscription.delete(subscription.id)
```

## Webhooks

### Overview
Webhooks provide real-time event notifications from Stripe to your application via HTTPS endpoints.

### Setup Process

#### 1. Create Webhook Handler
```python
import stripe
from django.http import HttpResponse
from django.views.decorators.csrf import csrf_exempt

@csrf_exempt
def webhook_handler(request):
    payload = request.body
    sig_header = request.META.get('HTTP_STRIPE_SIGNATURE')
    endpoint_secret = 'whsec_...'

    try:
        event = stripe.Webhook.construct_event(
            payload, sig_header, endpoint_secret
        )
    except stripe.error.SignatureVerificationError:
        return HttpResponse(status=400)

    # Handle specific event types
    if event['type'] == 'payment_intent.succeeded':
        handle_payment_success(event['data']['object'])
    elif event['type'] == 'payment_intent.payment_failed':
        handle_payment_failure(event['data']['object'])

    return HttpResponse(status=200)
```

#### 2. Register Webhook Endpoint
- Must be a public HTTPS URL
- Configure via Dashboard or API
- Select specific event types to receive
- Up to 16 endpoints per account

#### 3. Local Testing
```bash
# Install Stripe CLI
stripe listen --forward-to localhost:3000/webhooks/stripe

# Trigger test events
stripe trigger payment_intent.succeeded
```

### Important Events
- `payment_intent.succeeded`: Payment completed
- `payment_intent.payment_failed`: Payment failed
- `customer.subscription.created`: New subscription
- `customer.subscription.updated`: Subscription changed
- `invoice.payment_succeeded`: Invoice paid
- `invoice.payment_failed`: Invoice payment failed

### Security Best Practices
- Always verify webhook signatures
- Use raw request body for verification
- Return 200 status quickly
- Process events asynchronously when possible
- Handle duplicate events gracefully

## Error Handling

### Error Types

#### Card Errors
```python
try:
    stripe.PaymentIntent.create(...)
except stripe.error.CardError as e:
    # Payment was declined
    print(f"Card error: {e.user_message}")
```

#### Invalid Request Errors
```python
try:
    stripe.Customer.create(...)
except stripe.error.InvalidRequestError as e:
    # Invalid parameters
    print(f"Invalid request: {e.user_message}")
```

#### Authentication Errors
```python
try:
    stripe.Customer.list()
except stripe.error.AuthenticationError as e:
    # Invalid API key
    print(f"Authentication failed: {e.user_message}")
```

### Error Attributes
- `code`: Specific error code
- `message`: Error description
- `type`: Error category
- `param`: Parameter that caused the error
- `request_log_url`: Link to detailed request logs

### Best Practices
- Implement comprehensive error handling for all API calls
- Use test cards to simulate error scenarios
- Log errors with request IDs for debugging
- Provide user-friendly error messages
- Handle network timeouts and retries

## Best Practices

### Security
1. **API Key Management**
   - Never expose secret keys in client-side code
   - Use environment variables for key storage
   - Rotate keys regularly
   - Use restricted keys when possible

2. **Webhook Security**
   - Always verify webhook signatures
   - Use HTTPS endpoints only
   - Validate event data before processing

### Performance
1. **API Usage**
   - Use bulk operations when available
   - Implement proper caching strategies
   - Handle rate limits gracefully
   - Use pagination for large datasets

2. **Error Handling**
   - Implement exponential backoff for retries
   - Log errors with sufficient context
   - Monitor error rates and patterns

### Development
1. **Testing**
   - Use test mode for development
   - Test with various scenarios and error conditions
   - Implement automated tests for critical flows

2. **Monitoring**
   - Set up webhook monitoring
   - Track payment success/failure rates
   - Monitor API response times

## Code Examples

### Complete Payment Flow

#### Backend (Python/Django)
```python
import stripe
from django.conf import settings
from django.http import JsonResponse

stripe.api_key = settings.STRIPE_SECRET_KEY

def create_payment_intent(request):
    try:
        intent = stripe.PaymentIntent.create(
            amount=calculate_order_amount(request.POST.get('items')),
            currency='usd',
            customer=request.user.stripe_customer_id,
            metadata={
                'user_id': request.user.id,
                'order_id': request.POST.get('order_id')
            }
        )
        return JsonResponse({
            'client_secret': intent.client_secret
        })
    except Exception as e:
        return JsonResponse({'error': str(e)}, status=400)
```

#### Backend (Node.js/Express)
```javascript
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);

app.post('/create-payment-intent', async (req, res) => {
  try {
    const paymentIntent = await stripe.paymentIntents.create({
      amount: calculateOrderAmount(req.body.items),
      currency: 'usd',
      customer: req.body.customer_id,
      metadata: {
        order_id: req.body.order_id
      }
    });

    res.json({
      client_secret: paymentIntent.client_secret
    });
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
});
```

### Subscription Management
```python
def create_subscription(customer_id, price_id):
    try:
        subscription = stripe.Subscription.create(
            customer=customer_id,
            items=[{'price': price_id}],
            payment_behavior='default_incomplete',
            payment_settings={
                'save_default_payment_method': 'on_subscription'
            }
        )
        
        return {
            'subscription_id': subscription.id,
            'client_secret': subscription.latest_invoice.payment_intent.client_secret
        }
    except Exception as e:
        raise Exception(f"Subscription creation failed: {str(e)}")
```

### Customer with Saved Payment Method
```python
def setup_customer_with_payment_method(email, payment_method_id):
    try:
        # Create customer
        customer = stripe.Customer.create(email=email)
        
        # Attach payment method
        stripe.PaymentMethod.attach(
            payment_method_id,
            customer=customer.id
        )
        
        # Set as default
        stripe.Customer.modify(
            customer.id,
            invoice_settings={
                'default_payment_method': payment_method_id
            }
        )
        
        return customer
    except Exception as e:
        raise Exception(f"Customer setup failed: {str(e)}")
```

---

This comprehensive guide covers the essential aspects of Stripe backend integration. For the most up-to-date information, always refer to the official [Stripe documentation](https://docs.stripe.com/).