// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { KeyboardArrowRightRounded } from '@material-ui/icons';

import { numberWithCommas, parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import Button from '../../../components/button';
import NoStations from '../../../assets/images/noStations.svg';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';

const FailedStations = ({ createStationTrigger }) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };

    return (
        <div className="overview-wrapper failed-stations-container">
            <p className="overview-components-header">Stations</p>
            <div className="err-stations-list">
                {state?.monitor_data?.stations?.length > 0 ? (
                    <div className="coulmns-table">
                        <span style={{ width: '100px' }}>Name</span>
                        <span style={{ width: '200px' }}>Creation date</span>
                        <span style={{ width: '120px' }}>Total messages</span>
                        <span style={{ width: '120px' }}>Poison messages</span>
                        <span style={{ width: '120px' }}></span>
                    </div>
                ) : (
                    <div className="empty-stations-container">
                        <img src={NoStations} alt="no stations" onClick={() => createStationTrigger(true)} />
                        <p>No station Exsits</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder={'Create new station'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => createStationTrigger(true)}
                        />
                    </div>
                )}
                <div className="rows-wrapper">
                    {state?.monitor_data?.stations?.map((station, index) => {
                        return (
                            <div className="stations-row" key={index} onClick={() => goToStation(station.name)}>
                                <OverflowTip className="station-details" text={station.name} width={'100px'}>
                                    {station.name}
                                </OverflowTip>
                                <OverflowTip className="station-creation" text={parsingDate(station.creation_date)} width={'200px'}>
                                    {parsingDate(station.creation_date)}
                                </OverflowTip>
                                <span className="station-details centered">{numberWithCommas(station.total_messages)}</span>
                                <span className="station-details centered">{numberWithCommas(station.posion_messages)}</span>
                                <div className="link-wrapper">
                                    <div className="staion-link">
                                        <span>View Station</span>
                                        <KeyboardArrowRightRounded />
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default FailedStations;
