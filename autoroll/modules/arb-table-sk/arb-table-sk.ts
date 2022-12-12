/**
 * @module autoroll/modules/arb-table-sk
 * @description <h2><code>arb-table-sk</code></h2>
 *
 * <p>
 * This element displays the list of active Autorollers.
 * </p>
 */

import { html } from 'lit-html';
import { $$ } from 'common-sk/modules/dom';
import { define } from 'elements-sk/define';
import 'elements-sk/styles/table';

import { ElementSk } from '../../../infra-sk/modules/ElementSk';
import {
  AutoRollMiniStatus,
  AutoRollService,
  GetAutoRollService,
  GetRollersResponse,
  Mode,
} from '../rpc';
import { GetLastCheckInTime, LastCheckInSpan } from '../utils';

/**
 * hideOutdatedRollersThreshold is the threshold at which we'll stop displaying
 * a roller in the table by default. It can still be found if the user provides
 * their own filter or if they visit the roller's status page directly.
 */
const hideOutdatedRollersThreshold = 7 * 24 * 60 * 60 * 1000; // 7 days.

export class ARBTableSk extends ElementSk {
  private static template = (ele: ARBTableSk) => html`
  <div>
    Filter: <input id="filter" type="text" @input="${ele.updateFiltered
    }"></input>
  </div>
  <table>
    <tr>
      <th>Roller ID</th>
      <th>Current Mode</th>
      <th>Num Behind</th>
      <th>Num Failed</th>
      <th></th>
    </tr>
    ${ele.filtered.map(
      (st) => html`
        <tr>
          <td>
            <a href="/r/${st.rollerId}"
              >${st.childName} into ${st.parentName}</a
            >
          </td>
          <td class="${ele.modeClass(st.mode)}">${st.mode.toLowerCase()}</td>
          <td>${st.numBehind}</td>
          <td>${st.numFailed}</td>
          <td>${LastCheckInSpan(st)}</td>
        </tr>
      `,
    )}
  </table>
`;

  private rollers: AutoRollMiniStatus[] = [];

  private filtered: AutoRollMiniStatus[] = [];

  private rpc: AutoRollService;

  constructor() {
    super(ARBTableSk.template);
    this.rpc = GetAutoRollService(this);
  }

  connectedCallback() {
    super.connectedCallback();
    this.reload();
  }

  private modeClass(mode: Mode) {
    switch (mode) {
      case Mode.RUNNING:
        return "fg-running";
      case Mode.DRY_RUN:
        return "fg-dry-run";
      case Mode.STOPPED:
        return "fg-stopped";
      case Mode.OFFLINE:
        return "fg-offline";
    }
  }

  private reload() {
    this.rpc.getRollers({}).then((resp: GetRollersResponse) => {
      this.rollers = resp.rollers!;
      this.updateFiltered();
    });
  }

  private updateFiltered() {
    this.filtered = this.rollers;

    const filterInput = $$<HTMLInputElement>('#filter', this);
    if (!!filterInput && !!filterInput.value) {
      // If a filter was provided in the text box, use that.
      const regex = new RegExp(filterInput!.value);
      this.filtered = this.rollers.filter((st: AutoRollMiniStatus) => (
        st.rollerId.match(regex)
        || st.childName.match(regex)
        || st.parentName.match(regex)
      ));
    } else {
      // If no filter was provided, filter out any rollers which have not
      // checked in for longer than hideOutdatedRollersThreshold.
      this.filtered = this.rollers.filter((st: AutoRollMiniStatus) => {
        const lastCheckedIn = GetLastCheckInTime(st).getTime();
        const now = new Date().getTime()
        if (now - lastCheckedIn > hideOutdatedRollersThreshold) {
          return false;
        }
        return true;
      });
    }
    this._render();
  }
}

define('arb-table-sk', ARBTableSk);
