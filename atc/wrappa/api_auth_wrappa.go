package wrappa

import (
	"fmt"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/api/auth"
	"github.com/tedsuo/rata"
)

type APIAuthWrappa struct {
	checkPipelineAccessHandlerFactory   auth.CheckPipelineAccessHandlerFactory
	checkBuildReadAccessHandlerFactory  auth.CheckBuildReadAccessHandlerFactory
	checkBuildWriteAccessHandlerFactory auth.CheckBuildWriteAccessHandlerFactory
	checkWorkerTeamAccessHandlerFactory auth.CheckWorkerTeamAccessHandlerFactory
}

func NewAPIAuthWrappa(
	checkPipelineAccessHandlerFactory auth.CheckPipelineAccessHandlerFactory,
	checkBuildReadAccessHandlerFactory auth.CheckBuildReadAccessHandlerFactory,
	checkBuildWriteAccessHandlerFactory auth.CheckBuildWriteAccessHandlerFactory,
	checkWorkerTeamAccessHandlerFactory auth.CheckWorkerTeamAccessHandlerFactory,
) *APIAuthWrappa {
	return &APIAuthWrappa{
		checkPipelineAccessHandlerFactory:   checkPipelineAccessHandlerFactory,
		checkBuildReadAccessHandlerFactory:  checkBuildReadAccessHandlerFactory,
		checkBuildWriteAccessHandlerFactory: checkBuildWriteAccessHandlerFactory,
		checkWorkerTeamAccessHandlerFactory: checkWorkerTeamAccessHandlerFactory,
	}
}

func (wrappa *APIAuthWrappa) Wrap(handlers rata.Handlers) rata.Handlers {
	wrapped := rata.Handlers{}

	rejector := auth.UnauthorizedRejector{}

	for name, handler := range handlers {
		newHandler := handler

		switch name {
		// pipeline is public or authorized
		case atc.GetBuild,
			atc.BuildResources:
			newHandler = wrappa.checkBuildReadAccessHandlerFactory.AnyJobHandler(handler, rejector)

		// pipeline and job are public or authorized
		case atc.GetBuildPreparation,
			atc.BuildEvents,
			atc.GetBuildPlan,
			atc.ListBuildArtifacts:
			newHandler = wrappa.checkBuildReadAccessHandlerFactory.CheckIfPrivateJobHandler(handler, rejector)

			// resource belongs to authorized team
		case atc.AbortBuild,
			atc.SetBuildComment:
			newHandler = wrappa.checkBuildWriteAccessHandlerFactory.HandlerFor(handler, rejector)

		// requester is system, admin team, or worker owning team
		case atc.PruneWorker,
			atc.LandWorker,
			atc.RetireWorker,
			atc.ListDestroyingVolumes,
			atc.ListDestroyingContainers,
			atc.ReportWorkerContainers,
			atc.ReportWorkerVolumes:
			newHandler = wrappa.checkWorkerTeamAccessHandlerFactory.HandlerFor(handler, rejector)

		// pipeline is public or authorized
		case atc.GetPipeline,
			atc.GetJobBuild,
			atc.PipelineBadge,
			atc.JobBadge,
			atc.ListJobs,
			atc.GetJob,
			atc.ListJobBuilds,
			atc.ListPipelineBuilds,
			atc.GetResource,
			atc.ListBuildsWithVersionAsInput,
			atc.ListBuildsWithVersionAsOutput,
			atc.GetDownstreamResourceCausality,
			atc.GetUpstreamResourceCausality,
			atc.GetResourceVersion,
			atc.ListResources,
			atc.ListResourceTypes,
			atc.ListResourceVersions:
			newHandler = wrappa.checkPipelineAccessHandlerFactory.HandlerFor(handler, rejector)

		// authenticated
		case atc.ListWorkers,
			atc.RegisterWorker,
			atc.HeartbeatWorker,
			atc.DeleteWorker,
			atc.ListTeamBuilds,
			atc.GetUser:
			newHandler = auth.CheckAuthenticationHandler(handler, rejector)

		// unauthenticated / delegating to handler (validate token if provided)
		case atc.DownloadCLI,
			atc.CheckResourceWebHook,
			atc.TeamWebhook,
			atc.GetInfo,
			atc.ListTeams,
			atc.ListAllPipelines,
			atc.ListPipelines,
			atc.ListAllJobs,
			atc.ListAllResources,
			atc.ListBuilds,
			atc.MainJobBadge,
			atc.GetWall:
			newHandler = auth.CheckAuthenticationIfProvidedHandler(handler, rejector)

		// admin
		case atc.GetLogLevel,
			atc.DestroyTeam,
			atc.ListActiveUsersSince,
			atc.SetLogLevel,
			atc.GetInfoCreds,
			atc.SetWall,
			atc.ClearWall:
			newHandler = auth.CheckAdminHandler(handler, rejector)

		// authorized (requested team matches resource team and has required role, or is admin)
		case atc.GetTeam,
			atc.SetTeam,
			atc.RenameTeam,
			atc.ListContainers,
			atc.GetContainer,
			atc.HijackContainer,
			atc.ListVolumes,
			atc.CreateBuild,
			atc.CheckResource,
			atc.CheckResourceType,
			atc.CheckPrototype,
			atc.CreateJobBuild,
			atc.RerunJobBuild,
			atc.CreatePipelineBuild,
			atc.DeletePipeline,
			atc.DisableResourceVersion,
			atc.EnableResourceVersion,
			atc.PinResourceVersion,
			atc.UnpinResource,
			atc.SetPinCommentOnResource,
			atc.GetConfig,
			atc.GetCC,
			atc.GetVersionsDB,
			atc.ListJobInputs,
			atc.OrderPipelines,
			atc.OrderPipelinesWithinGroup,
			atc.PauseJob,
			atc.UnpauseJob,
			atc.PausePipeline,
			atc.UnpausePipeline,
			atc.RenamePipeline,
			atc.ExposePipeline,
			atc.HidePipeline,
			atc.SaveConfig,
			atc.SetTeamWebhook,
			atc.DestroyTeamWebhook,
			atc.ArchivePipeline,
			atc.ClearTaskCache,
			atc.ClearResourceCache,
			atc.CreateArtifact,
			atc.ScheduleJob,
			atc.GetArtifact:
			newHandler = auth.CheckAuthorizationHandler(handler, rejector)

		// think about it!
		default:
			panic(fmt.Sprintf("you missed a spot: %q", name))
		}

		wrapped[name] = newHandler
	}

	return wrapped
}
