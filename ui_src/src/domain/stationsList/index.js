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

import React, { useEffect, useContext, useState, useCallback, useRef } from 'react';
import { useHistory } from 'react-router-dom';

import { ApiEndpoints } from '../../const/apiEndpoints';
import StationBoxOverview from './stationBoxOverview';
import { httpRequest } from '../../services/http';
import Button from '../../components/button';
import { Context } from '../../hooks/store';
import SearchInput from '../../components/searchInput';
import pathDomains from '../../router';
import stationsIcon from '../../assets/images/stationIcon.svg';
import deleteWrapperIcon from '../../assets/images/deleteWrapperIcon.svg';
import searchIcon from '../../assets/images/searchIcon.svg';
import stationImg from '../../assets/images/stationsIconActive.svg';

import StationsInstructions from '../../components/stationsInstructions';
import Modal from '../../components/modal';
// import CreateStationDetails from '../../components/createStationDetails';
import CreateStationForm from '../../components/createStationForm';

import Loader from '../../components/loader';
import { stationFilterArray } from '../../services/valueConvertor';

const StationsList = () => {
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [modalDeleteIsOpen, modalDeleteFlip] = useState(false);
    const [stationsList, setStationList] = useState([]);
    const [filteredList, setFilteredList] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setisLoading] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);

    const createStationRef = useRef(null);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'stations' });
        getAllStations();
    }, []);

    useEffect(() => {
        if (searchInput.length >= 2) setFilteredList(stationsList.filter((station) => station.station.name.includes(searchInput)));
        else setFilteredList(stationsList);
    }, [searchInput]);

    useEffect(() => {
        if (searchInput !== '' && searchInput.length >= 2) {
            setFilteredList(stationsList.filter((station) => station.station.name.includes(searchInput)));
        } else setFilteredList(stationsList);
    }, [stationsList]);

    const getAllStations = async () => {
        setisLoading(true);
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_STATIONS}`);
            res.stations.sort((a, b) => new Date(b.station.creation_date) - new Date(a.station.creation_date));
            setStationList(res.stations);
            setFilteredList(res.stations);
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (err) {
            setisLoading(false);
            return;
        }
    };

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const handleRegisterToStation = useCallback(() => {
        state.socket?.emit('get_all_stations_data');
    }, [state.socket]);

    useEffect(() => {
        state.socket?.on(`stations_overview_data`, (data) => {
            data.sort((a, b) => new Date(b.station.creation_date) - new Date(a.station.creation_date));
            setStationList(data);
        });

        state.socket?.on('error', (error) => {
            history.push(pathDomains.overview);
        });

        setTimeout(() => {
            handleRegisterToStation();
        }, 1000);

        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const removeStation = async (stationName) => {
        try {
            await httpRequest('DELETE', ApiEndpoints.REMOVE_STATION, {
                station_name: stationName
            });
            setStationList(stationsList.filter((station) => station.station.name !== stationName));
        } catch (error) {
            return;
        }
    };

    const renderStationsOverview = () => {
        if (stationsList?.length > 0) {
            if (stationsList?.length <= 2) {
                return (
                    <div>
                        {filteredList?.map((station) => (
                            <StationBoxOverview
                                key={station.station.id}
                                isCheck={isCheck.includes(station.station.name)}
                                handleCheckedClick={handleCheckedClick}
                                station={station}
                                removeStation={() => removeStation(station.station.name)}
                            />
                        ))}
                        <StationsInstructions header="Add more stations" button="Add Station" newStation={() => modalFlip(true)} />
                    </div>
                );
            }
            return filteredList?.map((station) => (
                <StationBoxOverview
                    key={station.station.id}
                    isCheck={isCheck.includes(station.station.name)}
                    handleCheckedClick={handleCheckedClick}
                    station={station}
                    removeStation={() => removeStation(station.station.name)}
                />
            ));
        }
        return <StationsInstructions header="You don’t have any station yet?" button="Create New Station" image={stationsIcon} newStation={() => modalFlip(true)} />;
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(filteredList.map((li) => li.station.name));
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
        if (isCheck.length === 1 && !checked) {
            setIsCheckAll(false);
        }
    };

    const handleDeleteSelected = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_STATION, {
                station_names: isCheck
            });
            if (data) {
                setStationList(stationFilterArray(stationsList, isCheck));
                setIsCheck([]);
                setisLoading(false);
            }
        } catch (error) {
            setisLoading(false);
        }
        modalDeleteFlip(false);
    };

    return (
        <div className="stations-details-container">
            <div className="stations-details-header">
                <div className="header-wraper">
                    <label className="main-header-h1">
                        Stations <label className="length-list">{stationsList?.length > 0 && `(${stationsList?.length})`}</label>
                    </label>
                    {stationsList?.length > 0 && (
                        <div className="right-side">
                            <Button
                                width="131px"
                                height="34px"
                                placeholder={`Delete Selected (${isCheck?.length})`}
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="white"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                disabled={isCheck?.length === 0}
                                onClick={() => modalDeleteFlip(true)}
                            />

                            {filteredList?.length > 1 && (
                                <Button
                                    width="131px"
                                    height="34px"
                                    placeholder="Selected All"
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType="white"
                                    fontSize="12px"
                                    fontWeight="600"
                                    aria-haspopup="true"
                                    onClick={() => onCheckedAll()}
                                />
                            )}
                            <SearchInput
                                placeholder="Search Stations"
                                colorType="navy"
                                backgroundColorType="gray-dark"
                                width="288px"
                                height="34px"
                                borderColorType="none"
                                boxShadowsType="none"
                                borderRadiusType="circle"
                                iconComponent={<img src={searchIcon} />}
                                onChange={handleSearch}
                                value={searchInput}
                            />
                            <Button
                                className="modal-btn"
                                width="180px"
                                height="37px"
                                placeholder="Create New Station"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="14px"
                                fontWeight="bold"
                                aria-controls="usecse-menu"
                                aria-haspopup="true"
                                onClick={() => modalFlip(true)}
                            />
                        </div>
                    )}
                </div>
            </div>
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!isLoading && <div className="stations-content">{renderStationsOverview()}</div>}
            <div id="e2e-createstation-modal">
                <Modal
                    header={
                        <div className="modal-header">
                            <div className="header-img-container">
                                <img className="headerImage" src={stationImg} />
                            </div>
                            <p>Create new station</p>
                            <label>A station is a distributed unit that stores the produced data.</label>
                        </div>
                    }
                    height="460px"
                    width="540px"
                    rBtnText="Add"
                    lBtnText="Cancel"
                    lBtnClick={() => {
                        modalFlip(false);
                    }}
                    rBtnClick={() => {
                        createStationRef.current();
                    }}
                    clickOutside={() => modalFlip(false)}
                    open={modalIsOpen}
                    isLoading={creatingProsessd}
                >
                    <CreateStationForm createStationFormRef={createStationRef} handleClick={(e) => setCreatingProsessd(e)} />
                </Modal>
            </div>
            <Modal
                header={<img src={deleteWrapperIcon} />}
                width="520px"
                height="210px"
                displayButtons={false}
                clickOutside={() => modalDeleteFlip(false)}
                open={modalDeleteIsOpen}
            >
                <div className="roll-back-modal">
                    <p className="title">Are you sure you want to delete the selected stations?</p>
                    <p className="desc">Deleting these stations means they will be permanently deleted.</p>
                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => modalDeleteFlip(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Delete"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            loading={isLoading}
                            onClick={() => handleDeleteSelected()}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
};

export default StationsList;
