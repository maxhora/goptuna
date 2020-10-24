import { useEffect } from "react"
import { useRecoilState } from "recoil"
import { studyDetailsState } from "./state"
import { Action } from "./action"

export const useStudyDetail = (
  action: Action,
  studyId: number
): StudyDetail | null => {
  const [studyDetails, setStudyDetails] = useRecoilState<StudyDetails>(
    studyDetailsState
  )

  useEffect(() => {
    action.updateStudyDetail(studyId, studyDetails, setStudyDetails)
    const intervalId = setInterval(function () {
      action.updateStudyDetail(studyId, studyDetails, setStudyDetails)
    }, 10 * 1000)
    return () => clearInterval(intervalId)
  }, [])

  return studyDetails[studyId] || null
}
