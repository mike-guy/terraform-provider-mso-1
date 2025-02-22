package mso

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ciscoecosystem/mso-go-client/client"
	"github.com/ciscoecosystem/mso-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceMSOSchemaSiteVrfRegionCidrSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceMSOSchemaSiteVrfRegionCidrSubnetCreate,
		Read:   resourceMSOSchemaSiteVrfRegionCidrSubnetRead,
		Update: resourceMSOSchemaSiteVrfRegionCidrSubnetUpdate,
		Delete: resourceMSOSchemaSiteVrfRegionCidrSubnetDelete,

		Importer: &schema.ResourceImporter{
			State: resourceMSOSchemaSiteVrfRegionCidrSubnetImport,
		},

		SchemaVersion: version,

		Schema: (map[string]*schema.Schema{
			"schema_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"template_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"site_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"vrf_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"region_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"cidr_ip": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"ip": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"zone": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"usage": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
		}),
	}
}

func resourceMSOSchemaSiteVrfRegionCidrSubnetImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] %s: Beginning Import", d.Id())
	msoClient := m.(*client.Client)
	import_attribute := regexp.MustCompile("cidrIP/(.*)/subnet/(.*)")
	import_split := import_attribute.FindStringSubmatch(d.Id())
	get_attribute := strings.Split(d.Id(), "/")
	schemaId := get_attribute[0]
	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return nil, err
	}
	d.Set("schema_id", schemaId)
	count, err := cont.ArrayCount("sites")
	if err != nil {
		return nil, fmt.Errorf("No Sites found")
	}

	stateSite := get_attribute[2]
	found := false
	stateVrf := get_attribute[4]
	stateRegion := get_attribute[6]
	stateCidr := import_split[1]
	stateIp := import_split[2]

	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "sites")
		if err != nil {
			return nil, err
		}
		apiSite := models.StripQuotes(tempCont.S("siteId").String())

		if apiSite == stateSite {
			apiTemplate := models.StripQuotes(tempCont.S("templateName").String())
			vrfCount, err := tempCont.ArrayCount("vrfs")
			if err != nil {
				return nil, fmt.Errorf("Unable to get Vrf list")
			}
			for j := 0; j < vrfCount; j++ {
				vrfCont, err := tempCont.ArrayElement(j, "vrfs")
				if err != nil {
					return nil, err
				}
				apiVrfRef := models.StripQuotes(vrfCont.S("vrfRef").String())
				split := strings.Split(apiVrfRef, "/")
				apiVrf := split[6]
				if apiVrf == stateVrf {
					regionCount, err := vrfCont.ArrayCount("regions")
					if err != nil {
						return nil, fmt.Errorf("Unable to get Regions list")
					}
					for k := 0; k < regionCount; k++ {
						regionCont, err := vrfCont.ArrayElement(k, "regions")
						if err != nil {
							return nil, err
						}
						apiRegion := models.StripQuotes(regionCont.S("name").String())
						if apiRegion == stateRegion {
							cidrCount, err := regionCont.ArrayCount("cidrs")
							if err != nil {
								return nil, fmt.Errorf("Unable to get Cidr list")
							}
							for l := 0; l < cidrCount; l++ {
								cidrCont, err := regionCont.ArrayElement(l, "cidrs")
								if err != nil {
									return nil, err
								}
								apiCidr := models.StripQuotes(cidrCont.S("ip").String())
								log.Println("Current Cidr Ip", apiCidr)
								if apiCidr == stateCidr {
									subnetCount, err := cidrCont.ArrayCount("subnets")
									if err != nil {
										return nil, fmt.Errorf("Unable to get Subnet list")
									}
									for m := 0; m < subnetCount; m++ {
										subnetCont, err := cidrCont.ArrayElement(m, "subnets")
										if err != nil {
											return nil, err
										}
										apiIp := models.StripQuotes(subnetCont.S("ip").String())
										if apiIp == stateIp {
											d.SetId(apiIp)
											d.Set("ip", apiIp)
											d.Set("site_id", apiSite)
											d.Set("template_name", apiTemplate)
											d.Set("cidr_name", apiCidr)
											d.Set("vrf_name", apiVrf)
											d.Set("region_name", apiRegion)
											if subnetCont.Exists("zone") {
												d.Set("zone", models.StripQuotes(subnetCont.S("zone").String()))
											}
											if subnetCont.Exists("usage") {
												d.Set("usage", models.StripQuotes(subnetCont.S("usage").String()))
											}
											found = true
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if !found {
		d.SetId("")
		d.Set("schema_id", "")
		d.Set("site_id", "")
		d.Set("template_name", "")
		d.Set("region_name", "")
		d.Set("vrf_name", "")
		return nil, fmt.Errorf("Unable to find VRF Region Cidr Subnet %s", stateIp)

	}

	log.Printf("[DEBUG] %s: Import finished successfully", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceMSOSchemaSiteVrfRegionCidrSubnetCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Site Vrf Region Cidr Subnet: Beginning Creation")
	msoClient := m.(*client.Client)

	schemaId := d.Get("schema_id").(string)
	siteId := d.Get("site_id").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)
	regionName := d.Get("region_name").(string)
	cidrIp := d.Get("cidr_ip").(string)
	ip := d.Get("ip").(string)
	var zone, usage string
	if tempvar, ok := d.GetOk("zone"); ok {
		zone = tempvar.(string)
	}

	if tempvar, ok := d.GetOk("usage"); ok {
		usage = tempvar.(string)
	}

	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return err
	}
	count, err := cont.ArrayCount("sites")
	if err != nil {
		return fmt.Errorf("No Sites found")
	}

	cindex := -1
	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "sites")
		if err != nil {
			return err
		}
		apiSite := models.StripQuotes(tempCont.S("siteId").String())

		if apiSite == siteId {
			vrfCount, err := tempCont.ArrayCount("vrfs")
			if err != nil {
				return fmt.Errorf("Unable to get Vrf list")
			}
			for j := 0; j < vrfCount; j++ {
				vrfCont, err := tempCont.ArrayElement(j, "vrfs")
				if err != nil {
					return err
				}
				apiVrfRef := models.StripQuotes(vrfCont.S("vrfRef").String())
				split := strings.Split(apiVrfRef, "/")
				apiVrf := split[6]
				if apiVrf == vrfName {
					regionCount, err := vrfCont.ArrayCount("regions")
					if err != nil {
						return fmt.Errorf("Unable to get Regions list")
					}
					for k := 0; k < regionCount; k++ {
						regionCont, err := vrfCont.ArrayElement(k, "regions")
						if err != nil {
							return err
						}
						apiRegion := models.StripQuotes(regionCont.S("name").String())
						if apiRegion == regionName {
							cidrCount, err := regionCont.ArrayCount("cidrs")
							if err != nil {
								return fmt.Errorf("Unable to get Cidr list")
							}
							for l := 0; l < cidrCount; l++ {
								cidrCont, err := regionCont.ArrayElement(l, "cidrs")
								if err != nil {
									return err
								}
								apiCidr := models.StripQuotes(cidrCont.S("ip").String())
								log.Println("Current Cidr Ip", apiCidr)
								if apiCidr == cidrIp {
									cindex = l
									break
								}
							}
						}
					}
				}
			}
		}
	}

	path := fmt.Sprintf("/sites/%s-%s/vrfs/%s/regions/%s/cidrs/%v/subnets/-", siteId, templateName, vrfName, regionName, cindex)
	vrfRegionStruct := models.NewSchemaSiteVrfRegionCidrSubnet("add", path, ip, zone, usage)

	_, err1 := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaId), vrfRegionStruct)
	if err1 != nil {
		return err1
	}
	return resourceMSOSchemaSiteVrfRegionCidrSubnetRead(d, m)
}

func resourceMSOSchemaSiteVrfRegionCidrSubnetRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Beginning Read", d.Id())

	msoClient := m.(*client.Client)

	schemaId := d.Get("schema_id").(string)

	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return err
	}
	count, err := cont.ArrayCount("sites")
	if err != nil {
		return fmt.Errorf("No Sites found")
	}

	stateSite := d.Get("site_id").(string)
	found := false
	stateVrf := d.Get("vrf_name").(string)
	stateRegion := d.Get("region_name").(string)
	stateCidr := d.Get("cidr_ip").(string)
	stateIp := d.Get("ip").(string)

	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "sites")
		if err != nil {
			return err
		}
		apiSite := models.StripQuotes(tempCont.S("siteId").String())

		if apiSite == stateSite {
			apiTemplate := models.StripQuotes(tempCont.S("templateName").String())
			vrfCount, err := tempCont.ArrayCount("vrfs")
			if err != nil {
				return fmt.Errorf("Unable to get Vrf list")
			}
			for j := 0; j < vrfCount; j++ {
				vrfCont, err := tempCont.ArrayElement(j, "vrfs")
				if err != nil {
					return err
				}
				apiVrfRef := models.StripQuotes(vrfCont.S("vrfRef").String())
				split := strings.Split(apiVrfRef, "/")
				apiVrf := split[6]
				if apiVrf == stateVrf {
					regionCount, err := vrfCont.ArrayCount("regions")
					if err != nil {
						return fmt.Errorf("Unable to get Regions list")
					}
					for k := 0; k < regionCount; k++ {
						regionCont, err := vrfCont.ArrayElement(k, "regions")
						if err != nil {
							return err
						}
						apiRegion := models.StripQuotes(regionCont.S("name").String())
						if apiRegion == stateRegion {
							cidrCount, err := regionCont.ArrayCount("cidrs")
							if err != nil {
								return fmt.Errorf("Unable to get Cidr list")
							}
							for l := 0; l < cidrCount; l++ {
								cidrCont, err := regionCont.ArrayElement(l, "cidrs")
								if err != nil {
									return err
								}
								apiCidr := models.StripQuotes(cidrCont.S("ip").String())
								log.Println("Current Cidr Ip", apiCidr)
								if apiCidr == stateCidr {
									subnetCount, err := cidrCont.ArrayCount("subnets")
									if err != nil {
										return fmt.Errorf("Unable to get Subnet list")
									}
									for m := 0; m < subnetCount; m++ {
										subnetCont, err := cidrCont.ArrayElement(m, "subnets")
										if err != nil {
											return err
										}
										apiIp := models.StripQuotes(subnetCont.S("ip").String())

										if apiIp == stateIp {
											d.SetId(apiIp)
											d.Set("ip", apiIp)
											d.Set("site_id", apiSite)
											d.Set("template_name", apiTemplate)
											d.Set("cidr_name", apiCidr)
											d.Set("vrf_name", apiVrf)
											d.Set("region_name", apiRegion)
											if subnetCont.Exists("zone") {
												d.Set("zone", models.StripQuotes(subnetCont.S("zone").String()))
											}
											if subnetCont.Exists("usage") {
												d.Set("usage", models.StripQuotes(subnetCont.S("usage").String()))
											}
											found = true
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if !found {
		d.SetId("")
		d.Set("schema_id", "")
		d.Set("site_id", "")
		d.Set("template_name", "")
		d.Set("region_name", "")
		d.Set("vrf_name", "")

	}

	log.Printf("[DEBUG] %s: Read finished successfully", d.Id())
	return nil

}

func resourceMSOSchemaSiteVrfRegionCidrSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Site Vrf Region Cidr Subnet: Beginning Updation")
	msoClient := m.(*client.Client)

	schemaId := d.Get("schema_id").(string)
	siteId := d.Get("site_id").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)
	regionName := d.Get("region_name").(string)
	cidrIp := d.Get("cidr_ip").(string)
	ip := d.Get("ip").(string)

	var zone, usage string
	if tempvar, ok := d.GetOk("zone"); ok {
		zone = tempvar.(string)
	}

	if tempvar, ok := d.GetOk("usage"); ok {
		usage = tempvar.(string)
	}

	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return err
	}
	count, err := cont.ArrayCount("sites")
	if err != nil {
		return fmt.Errorf("No Sites found")
	}

	cindex := -1
	index := -1
	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "sites")
		if err != nil {
			return err
		}
		apiSite := models.StripQuotes(tempCont.S("siteId").String())

		if apiSite == siteId {
			vrfCount, err := tempCont.ArrayCount("vrfs")
			if err != nil {
				return fmt.Errorf("Unable to get Vrf list")
			}
			for j := 0; j < vrfCount; j++ {
				vrfCont, err := tempCont.ArrayElement(j, "vrfs")
				if err != nil {
					return err
				}
				apiVrfRef := models.StripQuotes(vrfCont.S("vrfRef").String())
				split := strings.Split(apiVrfRef, "/")
				apiVrf := split[6]
				if apiVrf == vrfName {
					regionCount, err := vrfCont.ArrayCount("regions")
					if err != nil {
						return fmt.Errorf("Unable to get Regions list")
					}
					for k := 0; k < regionCount; k++ {
						regionCont, err := vrfCont.ArrayElement(k, "regions")
						if err != nil {
							return err
						}
						apiRegion := models.StripQuotes(regionCont.S("name").String())
						if apiRegion == regionName {
							cidrCount, err := regionCont.ArrayCount("cidrs")
							if err != nil {
								return fmt.Errorf("Unable to get Cidr list")
							}
							for l := 0; l < cidrCount; l++ {
								cidrCont, err := regionCont.ArrayElement(l, "cidrs")
								if err != nil {
									return err
								}
								apiCidr := models.StripQuotes(cidrCont.S("ip").String())
								log.Println("Current Cidr Ip", apiCidr)
								if apiCidr == cidrIp {
									cindex = l
									subnetCount, err := cidrCont.ArrayCount("subnets")
									if err != nil {
										return fmt.Errorf("Unable to get Subnet list")
									}
									for m := 0; m < subnetCount; m++ {
										subnetCont, err := cidrCont.ArrayElement(m, "subnets")
										if err != nil {
											return err
										}
										apiIp := models.StripQuotes(subnetCont.S("ip").String())
										if apiIp == ip {
											index = m
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	path := fmt.Sprintf("/sites/%s-%s/vrfs/%s/regions/%s/cidrs/%v/subnets/%v", siteId, templateName, vrfName, regionName, cindex, index)
	vrfRegionStruct := models.NewSchemaSiteVrfRegionCidrSubnet("replace", path, ip, zone, usage)

	_, err1 := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaId), vrfRegionStruct)
	if err1 != nil {
		return err1
	}
	return resourceMSOSchemaSiteVrfRegionCidrSubnetRead(d, m)
}

func resourceMSOSchemaSiteVrfRegionCidrSubnetDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Site Vrf Region Cidr Subnet: Beginning Deletion")
	msoClient := m.(*client.Client)

	schemaId := d.Get("schema_id").(string)
	siteId := d.Get("site_id").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)
	regionName := d.Get("region_name").(string)
	cidrIp := d.Get("cidr_ip").(string)
	ip := d.Get("ip").(string)

	var zone, usage string
	if tempvar, ok := d.GetOk("zone"); ok {
		zone = tempvar.(string)
	}

	if tempvar, ok := d.GetOk("usage"); ok {
		usage = tempvar.(string)
	}

	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return err
	}
	count, err := cont.ArrayCount("sites")
	if err != nil {
		return fmt.Errorf("No Sites found")
	}

	cindex := -1
	index := -1
	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "sites")
		if err != nil {
			return err
		}
		apiSite := models.StripQuotes(tempCont.S("siteId").String())

		if apiSite == siteId {
			vrfCount, err := tempCont.ArrayCount("vrfs")
			if err != nil {
				return fmt.Errorf("Unable to get Vrf list")
			}
			for j := 0; j < vrfCount; j++ {
				vrfCont, err := tempCont.ArrayElement(j, "vrfs")
				if err != nil {
					return err
				}
				apiVrfRef := models.StripQuotes(vrfCont.S("vrfRef").String())
				split := strings.Split(apiVrfRef, "/")
				apiVrf := split[6]
				if apiVrf == vrfName {
					regionCount, err := vrfCont.ArrayCount("regions")
					if err != nil {
						return fmt.Errorf("Unable to get Regions list")
					}
					for k := 0; k < regionCount; k++ {
						regionCont, err := vrfCont.ArrayElement(k, "regions")
						if err != nil {
							return err
						}
						apiRegion := models.StripQuotes(regionCont.S("name").String())
						if apiRegion == regionName {
							cidrCount, err := regionCont.ArrayCount("cidrs")
							if err != nil {
								return fmt.Errorf("Unable to get Cidr list")
							}
							for l := 0; l < cidrCount; l++ {
								cidrCont, err := regionCont.ArrayElement(l, "cidrs")
								if err != nil {
									return err
								}
								apiCidr := models.StripQuotes(cidrCont.S("ip").String())
								log.Println("Current Cidr Ip", apiCidr)
								if apiCidr == cidrIp {
									cindex = l
									subnetCount, err := cidrCont.ArrayCount("subnets")
									if err != nil {
										return fmt.Errorf("Unable to get Subnet list")
									}
									for m := 0; m < subnetCount; m++ {
										subnetCont, err := cidrCont.ArrayElement(m, "subnets")
										if err != nil {
											return err
										}
										apiIp := models.StripQuotes(subnetCont.S("ip").String())
										if apiIp == ip {
											index = m
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if index == -1 {
		d.SetId("")
		return nil
	}

	path := fmt.Sprintf("/sites/%s-%s/vrfs/%s/regions/%s/cidrs/%v/subnets/%v", siteId, templateName, vrfName, regionName, cindex, index)
	vrfRegionStruct := models.NewSchemaSiteVrfRegionCidrSubnet("remove", path, ip, zone, usage)

	response, err1 := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaId), vrfRegionStruct)

	// Ignoring Error with code 141: Resource Not Found when deleting
	if err1 != nil && !(response.Exists("code") && response.S("code").String() == "141") {
		return err1
	}
	d.SetId("")
	return nil
}
