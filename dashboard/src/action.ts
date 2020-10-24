import { SetterOrUpdater, useRecoilState } from "recoil"
import { useSnackbar } from "notistack"
import {
  getStudyDetailAPI,
  getStudySummariesAPI,
  createNewStudyAPI,
} from "./apiClient"
import { studySummariesState } from "./state"

export const actionCreator = () => {
  const { enqueueSnackbar } = useSnackbar()
  const [studySummaries, setStudySummaries] = useRecoilState<StudySummary[]>(
    studySummariesState
  )

  const updateStudySummaries = (successMsg?: string) => {
    getStudySummariesAPI()
      .then((studySummaries: StudySummary[]) => {
        setStudySummaries(studySummaries)

        if (successMsg) {
          enqueueSnackbar(successMsg, { variant: "success" })
        }
      })
      .catch((err) => {
        enqueueSnackbar(`Failed to fetch study list.`, {
          variant: "error",
        })
        console.log(err)
      })
  }

  const updateStudyDetail = (
    studyId: number,
    oldVal: StudyDetails,
    setter: SetterOrUpdater<StudyDetails>
  ) => {
    getStudyDetailAPI(studyId)
      .then((study) => {
        let newVal = Object.assign({}, oldVal)
        newVal[studyId] = study
        setter(newVal)
      })
      .catch((err) => {
        enqueueSnackbar(`Failed to fetch study (id=${studyId})`, {
          variant: "error",
        })
        console.log(err)
      })
  }

  const createNewStudy = (studyName: string, direction: StudyDirection) => {
    createNewStudyAPI(studyName, direction)
      .then((study_summary) => {
        const newVal = [...studySummaries, study_summary]
        setStudySummaries(newVal)
        enqueueSnackbar(`Success to create a study (study_name=${studyName})`, {
          variant: "success",
        })
      })
      .catch((err) => {
        enqueueSnackbar(`Failed to create a study (study_name=${studyName})`, {
          variant: "error",
        })
        console.log(err)
      })
  }

  return {
    updateStudyDetail,
    updateStudySummaries,
    createNewStudy,
  }
}

export type Action = ReturnType<typeof actionCreator>
