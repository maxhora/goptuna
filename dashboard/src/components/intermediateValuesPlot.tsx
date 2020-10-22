import * as plotly from "plotly.js-dist"
import React, { FC, useEffect } from "react"

const plotDomId = "intermediate-values-plot"

export const IntermediateValuesPlot: FC<{
  trials: Trial[]
}> = ({ trials = [] }) => {
  useEffect(() => {
    plotIntermediateValue(trials)
  }, [trials])
  return <div id={plotDomId} />
}

const plotIntermediateValue = (trials: Trial[]) => {
  if (document.getElementById(plotDomId) === null) {
    return
  }

  const layout: Partial<plotly.Layout> = {
    title: "Intermediate values",
    margin: {
      l: 50,
      r: 50,
      b: 0,
    },
  }
  if (trials.length === 0) {
    plotly.react(plotDomId, [], layout)
    return
  }

  let filteredTrials = trials.filter(
    (t) => t.state === TrialState.Complete || t.state === TrialState.Pruned
  )
  const plotData: Partial<plotly.PlotData>[] = filteredTrials.map((trial) => {
    return {
      x: trial.intermediate_values.map((iv) => iv.step),
      y: trial.intermediate_values.map((iv) => iv.value),
      mode: "lines+markers",
      type: "scatter",
      name: `trial #${trial.number}`,
    }
  })
  plotly.react(plotDomId, plotData, layout)
}
