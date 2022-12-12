import './index';
import { ClusterSummary2Sk } from './cluster-summary2-sk';
import {
  FullSummary, ClusterSummary, TriageStatus, Alert, FrameResponse,
} from '../json';

Date.now = () => Date.parse('2020-03-22T00:00:00.000Z');

// Handle the sk namespace attached to window.
declare global {
  interface Window {
    Login: any;
  }
}

window.Login = Promise.resolve({
  Email: 'user@google.com',
  LoginURL: 'https://accounts.google.com/',
});

ClusterSummary2Sk.lookupCids = () => new Promise((resolve) => {
  resolve({
    commitSlice: [
      {
        hash: '566ac1e41bd7953a7179da9832bc883e1c758fac',
        offset: 24748,
        author: 'msarett@google.com',
        message: 'Safely handle unsupported color xforms in SkCodec',
        url:
          'https://skia.googlesource.com/skia/+show/313c4635e3f1005e6807f5b0ad52805f30902d66',
        ts: 1476984695,
      },
    ],
    logEntry: '',
  });
});

const alert: Alert = {
  id_as_string: '-1',
  display_name: '',
  radius: 6,
  query: 'config=565',
  k: 0,
  algo: 'stepfit',
  interesting: 0.05,
  sparse: false,
  step: '',
  alert: '',
  bug_uri_template: '',
  state: 'ACTIVE',
  owner: '',
  step_up_only: false,
  direction: 'BOTH',
  group_by: '',
  minimum_num: 0,
  category: '',
};

const summary: ClusterSummary = {
  centroid: [
    -1.0826576,
    0.33417022,
    0.8747909,
    0.11694965,
    0.76775414,
    -0.21376616,
    0.026059598,
    -0.08791064,
    0.13508978,
    -0.38292113,
    -0.4874483,
  ],
  shortcut: 'X123',
  param_summaries2: [
    { value: 'arch=arm', percent: 40 },
    { value: 'arch=arm64', percent: 30 },
    { value: 'arch=x86', percent: 20 },
    { value: 'bench_type=skandroidcodec', percent: 10 },
    { value: 'arch=x86_64', percent: 1 },
  ],
  step_fit: {
    least_squares: 0.12262289,
    turning_point: 1,
    step_size: -1.1909344,
    regression: -200.712171,
    status: 'Low',
  },
  step_point: {
    offset: 24745,
    timestamp: 1476983221,
  },
  num: 4,
  ts: new Date().toISOString(),
};

const frame: FrameResponse = {
  dataframe: {
    traceset: {},
    header: [
      { offset: 24744, timestamp: 1476982874 },
      { offset: 24745, timestamp: 1476983221 },
      { offset: 24746, timestamp: 1476983487 },
      { offset: 24747, timestamp: 1476983833 },
      { offset: 24748, timestamp: 1476984695 },
      { offset: 24749, timestamp: 1476985138 },
      { offset: 24750, timestamp: 1476985844 },
      { offset: 24751, timestamp: 1476986630 },
      { offset: 24752, timestamp: 1476986672 },
      { offset: 24753, timestamp: 1476986679 },
      { offset: 24754, timestamp: 1476987166 },
    ],
    paramset: {
      arch: ['arm', 'arm64', 'x86', 'x86_64'],
      bench_type: ['skandroidcodec'],
      compiler: ['Clang', 'GCC', 'MSVC'],
      config: ['nonrendering'],
      cpu_or_gpu: ['CPU'],
    },
    skip: 0,
  },
  skps: [],
  msg: '',
  display_mode: 'display_plot',
};

const triage: TriageStatus = {
  status: 'untriaged',
  message: 'Nothing to see here.',
};

const fullSummary: FullSummary = {
  summary,
  triage,
  frame,
};

const cluster = document.querySelector<ClusterSummary2Sk>(
  'cluster-summary2-sk.cluster',
)!;
cluster.full_summary = fullSummary;
cluster.alert = alert;

const alert2 = JSON.parse(JSON.stringify(alert)) as Alert;
alert2.step = 'mannwhitneyu';
const summary2 = JSON.parse(JSON.stringify(fullSummary)) as FullSummary;
summary2.summary.step_fit!.status = 'High';
summary2.summary.step_fit!.regression = 0.06732;
summary2.summary.step_fit!.least_squares = 12;
const nostatus = document.querySelector<ClusterSummary2Sk>(
  'cluster-summary2-sk.nostatus',
)!;
nostatus.full_summary = summary2;
nostatus.triage = triage;
nostatus.alert = alert2;

document.body.addEventListener('triaged', (e) => {
  document.querySelector('code.events')!.textContent = JSON.stringify(
    (e as CustomEvent).detail,
    null,
    ' ',
  );
});
document.body.addEventListener('open-keys', (e) => {
  document.querySelector('code.events')!.textContent = JSON.stringify(
    (e as CustomEvent).detail,
    null,
    ' ',
  );
});
