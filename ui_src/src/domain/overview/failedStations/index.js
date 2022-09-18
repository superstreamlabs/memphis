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
import { Link } from 'react-router-dom';

import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';

const FailedStations = () => {
    const [state, dispatch] = useContext(Context);
    return (
        <div className="overview-wrapper failed-stations-container">
            <p className="overview-components-header" id="e2e-overview-station-list">
                Stations
            </p>
            <div className="err-stations-list">
                <div className="coulmns-table">
                    <span style={{ width: '100px' }}>Name</span>
                    <span style={{ width: '200px' }}>Creation date</span>
                    <span style={{ width: '100px' }}>Created by</span>
                    <span style={{ width: '100px' }}></span>
                </div>
                <div className="rows-wrapper">
                    {state?.monitor_data?.stations?.map((station, index) => {
                        return (
                            <div className="stations-row" key={index}>
                                <OverflowTip text={station.name} width={'100px'}>
                                    {station.name}
                                </OverflowTip>
                                <OverflowTip text={parsingDate(station.creation_date)} width={'200px'}>
                                    {parsingDate(station.creation_date)}
                                </OverflowTip>
                                <OverflowTip text={station.created_by_user} width={'100px'}>
                                    {station.created_by_user}
                                </OverflowTip>
                                <Link style={{ cursor: 'pointer' }} to={`${pathDomains.stations}/${station.name}`}>
                                    <span className="link-row" style={{ width: '100px' }}>
                                        Go to station
                                    </span>
                                </Link>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default FailedStations;
