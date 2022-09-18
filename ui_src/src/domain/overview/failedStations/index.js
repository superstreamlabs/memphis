// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useContext } from 'react';
import { Link, useHistory } from 'react-router-dom';

import { numberWithCommas, parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import staionLink from '../../../assets/images/staionLink.svg';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';

const FailedStations = () => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };

    return (
        <div className="overview-wrapper failed-stations-container">
            <p className="overview-components-header" id="e2e-overview-station-list">
                Stations
            </p>
            <div className="err-stations-list">
                <div className="coulmns-table">
                    <span style={{ width: '100px' }}>Name</span>
                    <span style={{ width: '200px' }}>Creation date</span>
                    <span style={{ width: '120px' }}>Total messages</span>
                    <span style={{ width: '120px' }}>Poison messages</span>
                    <span style={{ width: '120px' }}></span>
                </div>
                <div className="rows-wrapper">
                    {state?.monitor_data?.stations?.map((station, index) => {
                        return (
                            <div className="stations-row" key={index}>
                                <OverflowTip className="station-details" text={station.name} width={'100px'}>
                                    {station.name}
                                </OverflowTip>
                                <OverflowTip className="station-creation" text={parsingDate(station.creation_date)} width={'200px'}>
                                    {parsingDate(station.creation_date)}
                                </OverflowTip>
                                <span className="station-details" style={{ width: '120px' }}>
                                    {numberWithCommas(station.total_messages)}
                                </span>
                                <span className="station-details" style={{ width: '120px' }}>
                                    {numberWithCommas(station.posion_messages)}
                                </span>
                                <div className="staion-link" onClick={() => goToStation(station.name)}>
                                    <span>View Station</span>
                                    <img src={staionLink} />
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
