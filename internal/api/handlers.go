package api

// TODO: Implement public API handlers:
//
// --- Instances ---
//
// func (s *Server) HandleListInstances(w http.ResponseWriter, r *http.Request)
//   - Extract org_id from auth claims
//   - Optional ?status= query filter
//   - Query db.ListInstances(ctx, orgID, status)
//   - Return JSON array of instance responses (strip upstream fields)
//
// func (s *Server) HandleCreateInstance(w http.ResponseWriter, r *http.Request)
//   - Decode CreateInstanceRequest from JSON body
//   - Validate request fields
//   - Check billing status via billing.CheckBillingStatus()
//   - Call provision.Engine.Provision()
//   - Return InstanceResponse with hostname, ssh_command, etc.
//
// func (s *Server) HandleGetInstance(w http.ResponseWriter, r *http.Request)
//   - Extract {id} from path
//   - Verify org ownership
//   - Return instance details (strip upstream fields)
//
// func (s *Server) HandleDeleteInstance(w http.ResponseWriter, r *http.Request)
//   - Extract {id} from path
//   - Verify org ownership
//   - Call provision.Engine.Terminate()
//   - Return success
//
// func (s *Server) HandleInstanceStatus(w http.ResponseWriter, r *http.Request)
//   - Extract {id} from path
//   - Return current status from DB (+ optional live check)
//
// --- GPU Availability ---
//
// func (s *Server) HandleListAvailable(w http.ResponseWriter, r *http.Request)
//   - Parse query params: type, tier, region, sort_by
//   - Read Redis availability cache
//   - Filter and sort offerings
//   - Strip provider field (customer must not see upstream source)
//   - Return JSON array
//
// --- SSH Keys ---
//
// func (s *Server) HandleListSSHKeys(w http.ResponseWriter, r *http.Request)
// func (s *Server) HandleCreateSSHKey(w http.ResponseWriter, r *http.Request)
// func (s *Server) HandleDeleteSSHKey(w http.ResponseWriter, r *http.Request)
//
// --- Billing ---
//
// func (s *Server) HandleGetUsage(w http.ResponseWriter, r *http.Request)
// func (s *Server) HandleGetInvoices(w http.ResponseWriter, r *http.Request)
// func (s *Server) HandleStripeWebhook(w http.ResponseWriter, r *http.Request)
//
// --- Internal (cloud-init callback) ---
//
// func (s *Server) HandleInstanceReady(w http.ResponseWriter, r *http.Request)
//   - Called by cloud-init when instance boots successfully
//   - Update instance status to "running", set billing_start
//
// func (s *Server) HandleInstanceHealth(w http.ResponseWriter, r *http.Request)
//   - Called by instance health pings
//   - Update last_seen timestamp
