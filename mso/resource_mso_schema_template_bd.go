package mso

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ciscoecosystem/mso-go-client/client"
	"github.com/ciscoecosystem/mso-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceMSOTemplateBD() *schema.Resource {
	return &schema.Resource{
		Create: resourceMSOTemplateBDCreate,
		Read:   resourceMSOTemplateBDRead,
		Update: resourceMSOTemplateBDUpdate,
		Delete: resourceMSOTemplateBDDelete,

		Importer: &schema.ResourceImporter{
			State: resourceMSOTemplateBDImport,
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
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"display_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"intersite_bum_traffic": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"optimize_wan_bandwidth": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"layer2_stretch": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"layer3_multicast": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"layer2_unknown_unicast": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"unknown_multicast_flooding": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"flood",
					"optimized_flooding",
				}, false),
			},
			"multi_destination_flooding": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"flood_in_bd",
					"drop",
					"flood_in_encap",
				}, false),
			},
			"ipv6_unknown_multicast_flooding": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"flood",
					"optimized_flooding",
				}, false),
			},
			"arp_flooding": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"virtual_mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"unicast_routing": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"vrf_name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"vrf_schema_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vrf_template_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dhcp_policy": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"dhcp_option_policy_name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"dhcp_option_policy_version": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		}),
	}
}

func resourceMSOTemplateBDImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] %s: Beginning Import", d.Id())

	msoClient := m.(*client.Client)
	get_attribute := strings.Split(d.Id(), "/")
	schemaId := get_attribute[0]
	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return nil, err
	}
	count, err := cont.ArrayCount("templates")
	if err != nil {
		return nil, fmt.Errorf("No Template found")
	}
	stateTemplate := get_attribute[2]
	found := false
	stateBD := get_attribute[4]
	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "templates")
		if err != nil {
			return nil, err
		}
		apiTemplate := models.StripQuotes(tempCont.S("name").String())

		if apiTemplate == stateTemplate {
			bdCount, err := tempCont.ArrayCount("bds")
			if err != nil {
				return nil, fmt.Errorf("Unable to get BD list")
			}
			for j := 0; j < bdCount; j++ {
				bdCont, err := tempCont.ArrayElement(j, "bds")
				if err != nil {
					return nil, err
				}
				apiBD := models.StripQuotes(bdCont.S("name").String())
				if apiBD == stateBD {
					d.SetId(apiBD)
					d.Set("name", apiBD)
					d.Set("schema_id", schemaId)
					d.Set("template_name", apiTemplate)
					d.Set("display_name", models.StripQuotes(bdCont.S("displayName").String()))
					d.Set("layer2_unknown_unicast", models.StripQuotes(bdCont.S("l2UnknownUnicast").String()))
					d.Set("unknown_multicast_flooding", models.StripQuotes(bdCont.S("unkMcastAct").String()))
					d.Set("multi_destination_flooding", models.StripQuotes(bdCont.S("multiDstPktAct").String()))
					d.Set("ipv6_unknown_multicast_flooding", models.StripQuotes(bdCont.S("v6unkMcastAct").String()))
					d.Set("virtual_mac_address", models.StripQuotes(bdCont.S("vmac").String()))
					if bdCont.Exists("intersiteBumTrafficAllow") {
						d.Set("intersite_bum_traffic", bdCont.S("intersiteBumTrafficAllow").Data().(bool))
					}

					if bdCont.Exists("optimizeWanBandwidth") {
						d.Set("optimize_wan_bandwidth", bdCont.S("optimizeWanBandwidth").Data().(bool))
					}

					if bdCont.Exists("l3MCast") {
						d.Set("layer3_multicast", bdCont.S("l3MCast").Data().(bool))
					}

					if bdCont.Exists("l2Stretch") {
						d.Set("layer2_stretch", bdCont.S("l2Stretch").Data().(bool))
					}
					if bdCont.Exists("arpFlood") {
						d.Set("arp_flooding", bdCont.S("arpFlood").Data().(bool))
					}
					if bdCont.Exists("unicastRouting") {
						d.Set("unicast_routing", bdCont.S("unicastRouting").Data().(bool))
					}

					vrfRef := models.StripQuotes(bdCont.S("vrfRef").String())
					re := regexp.MustCompile("/schemas/(.*)/templates/(.*)/vrfs/(.*)")
					match := re.FindStringSubmatch(vrfRef)
					d.Set("vrf_name", match[3])
					d.Set("vrf_schema_id", match[1])
					d.Set("vrf_template_name", match[2])

					dhcpPolMap := make(map[string]interface{})
					dhcpPolMap["name"] = models.StripQuotes(bdCont.S("dhcpLabel", "name").String())
					if dhcpPolMap["name"] == "{}" {
						dhcpPolMap = nil
					} else {
						dhcpPolMap["version"] = models.StripQuotes(bdCont.S("dhcpLabel", "version").String())
						if dhcpPolMap["version"] == "{}" {
							dhcpPolMap["version"] = nil
						}
						if bdCont.Exists("dhcpLabel", "dhcpOptionLabel") {
							dhcpPolMap["dhcp_option_policy_name"] = models.StripQuotes(bdCont.S("dhcpLabel", "dhcpOptionLabel", "name").String())
							dhcpPolMap["dhcp_option_policy_version"] = models.StripQuotes(bdCont.S("dhcpLabel", "dhcpOptionLabel", "version").String())
							if dhcpPolMap["dhcp_option_policy_name"] == "{}" {
								dhcpPolMap["dhcp_option_policy_name"] = nil
							}
							if dhcpPolMap["dhcp_option_policy_version"] == "{}" {
								dhcpPolMap["dhcp_option_policy_version"] = nil
							}
						}
					}
					d.Set("dhcp_policy", dhcpPolMap)
					found = true
					break
				}

			}
		}

	}

	if !found {
		d.SetId("")
	}

	log.Printf("[DEBUG] %s: Import finished successfully", d.Id())
	return []*schema.ResourceData{d}, nil

}

func resourceMSOTemplateBDCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Template BD: Beginning Creation")
	msoClient := m.(*client.Client)

	schemaID := d.Get("schema_id").(string)
	name := d.Get("name").(string)
	displayName := d.Get("display_name").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)

	var intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, unicast_routing, arp_flooding bool
	var layer2_unknown_unicast, vrf_schema_id, vrf_template_name, virtual_mac_address, ipv6_unknown_multicast_flooding, multi_destination_flooding, unknown_multicast_flooding string

	if tempVar, ok := d.GetOk("intersite_bum_traffic"); ok {
		intersite_bum_traffic = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("optimize_wan_bandwidth"); ok {
		optimize_wan_bandwidth = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer2_stretch"); ok {
		layer2_stretch = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer3_multicast"); ok {
		layer3_multicast = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer2_unknown_unicast"); ok {
		layer2_unknown_unicast = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unknown_multicast_flooding"); ok {
		unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("multi_destination_flooding"); ok {
		multi_destination_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("ipv6_unknown_multicast_flooding"); ok {
		ipv6_unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unicast_routing"); ok {
		unicast_routing = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("virtual_mac_address"); ok {
		virtual_mac_address = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("arp_flooding"); ok {
		arp_flooding = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("vrf_schema_id"); ok {
		vrf_schema_id = tempVar.(string)
	} else {
		vrf_schema_id = schemaID
	}
	if tempVar, ok := d.GetOk("vrf_template_name"); ok {
		vrf_template_name = tempVar.(string)
	} else {
		vrf_template_name = templateName
	}
	var dhcpPolMap map[string]interface{}
	if tempVar, ok := d.GetOk("dhcp_policy"); ok {
		dhcp_policy := tempVar.(map[string]interface{})
		dhcpPolMap = make(map[string]interface{})
		dhcpPolMap["name"] = dhcp_policy["name"]
		version, err := strconv.Atoi(dhcp_policy["version"].(string))
		if err != nil {
			return err
		}
		dhcpPolMap["version"] = version

		optionName := dhcp_policy["dhcp_option_policy_name"]
		optionVersion := dhcp_policy["dhcp_option_policy_version"]
		if optionName != nil {
			if optionVersion != nil {
				dhcpOptionMap := make(map[string]interface{})
				dhcpOptionMap["name"] = optionName
				ver, err := strconv.Atoi(optionVersion.(string))
				if err != nil {
					return err
				}
				dhcpOptionMap["version"] = ver
				dhcpPolMap["dhcpOptionLabel"] = dhcpOptionMap
			} else {
				return fmt.Errorf("dhcp_option_policy_version is required with dhcp_option_policy_name")
			}
		}
	} else {
		dhcpPolMap = nil
	}

	vrfRefMap := make(map[string]interface{})
	vrfRefMap["schemaId"] = vrf_schema_id
	vrfRefMap["templateName"] = vrf_template_name
	vrfRefMap["vrfName"] = vrfName
	path := fmt.Sprintf("/templates/%s/bds/-", templateName)
	bdStruct := models.NewTemplateBD("add", path, name, displayName, layer2_unknown_unicast, unknown_multicast_flooding, multi_destination_flooding, ipv6_unknown_multicast_flooding, virtual_mac_address, intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, arp_flooding, unicast_routing, vrfRefMap, dhcpPolMap)

	_, err := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaID), bdStruct)

	if err != nil {
		return err
	}
	return resourceMSOTemplateBDRead(d, m)
}

func resourceMSOTemplateBDRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Beginning Read", d.Id())

	msoClient := m.(*client.Client)

	schemaId := d.Get("schema_id").(string)

	cont, err := msoClient.GetViaURL(fmt.Sprintf("api/v1/schemas/%s", schemaId))
	if err != nil {
		return err
	}
	count, err := cont.ArrayCount("templates")
	if err != nil {
		return fmt.Errorf("No Template found")
	}
	stateTemplate := d.Get("template_name").(string)
	found := false
	stateBD := d.Get("name")
	for i := 0; i < count; i++ {
		tempCont, err := cont.ArrayElement(i, "templates")
		if err != nil {
			return err
		}
		apiTemplate := models.StripQuotes(tempCont.S("name").String())

		if apiTemplate == stateTemplate {
			bdCount, err := tempCont.ArrayCount("bds")
			if err != nil {
				return fmt.Errorf("Unable to get BD list")
			}
			for j := 0; j < bdCount; j++ {
				bdCont, err := tempCont.ArrayElement(j, "bds")
				if err != nil {
					return err
				}

				apiBD := models.StripQuotes(bdCont.S("name").String())
				if apiBD == stateBD {
					d.SetId(apiBD)
					d.Set("name", apiBD)
					d.Set("schema_id", schemaId)
					d.Set("template_name", apiTemplate)
					d.Set("display_name", models.StripQuotes(bdCont.S("displayName").String()))
					d.Set("layer2_unknown_unicast", models.StripQuotes(bdCont.S("l2UnknownUnicast").String()))
					d.Set("unknown_multicast_flooding", models.StripQuotes(bdCont.S("unkMcastAct").String()))
					d.Set("multi_destination_flooding", models.StripQuotes(bdCont.S("multiDstPktAct").String()))
					d.Set("ipv6_unknown_multicast_flooding", models.StripQuotes(bdCont.S("v6unkMcastAct").String()))
					d.Set("virtual_mac_address", models.StripQuotes(bdCont.S("vmac").String()))
					if bdCont.Exists("intersiteBumTrafficAllow") {
						d.Set("intersite_bum_traffic", bdCont.S("intersiteBumTrafficAllow").Data().(bool))
					}

					if bdCont.Exists("optimize_wan_bandwidth") {
						d.Set("optimize_wan_bandwidth", bdCont.S("optimizeWanBandwidth").Data().(bool))
					}

					if bdCont.Exists("layer3_multicast") {
						d.Set("layer3_multicast", bdCont.S("l3MCast").Data().(bool))
					}

					if bdCont.Exists("layer2_stretch") {
						d.Set("layer2_stretch", bdCont.S("l2Stretch").Data().(bool))
					}
					if bdCont.Exists("arp_flooding") {
						d.Set("arp_flooding", bdCont.S("arpFlood").Data().(bool))
					}
					if bdCont.Exists("unicast_routing") {
						d.Set("unicast_routing", bdCont.S("unicastRouting").Data().(bool))
					}

					vrfRef := models.StripQuotes(bdCont.S("vrfRef").String())
					re := regexp.MustCompile("/schemas/(.*)/templates/(.*)/vrfs/(.*)")
					match := re.FindStringSubmatch(vrfRef)
					d.Set("vrf_name", match[3])
					d.Set("vrf_schema_id", match[1])
					d.Set("vrf_template_name", match[2])

					dhcpPolMap := make(map[string]interface{})
					dhcpPolMap["name"] = models.StripQuotes(bdCont.S("dhcpLabel", "name").String())
					if dhcpPolMap["name"] == "{}" {
						dhcpPolMap = nil
					} else {
						dhcpPolMap["version"] = models.StripQuotes(bdCont.S("dhcpLabel", "version").String())
						if dhcpPolMap["version"] == "{}" {
							dhcpPolMap["version"] = nil
						}
						if bdCont.Exists("dhcpLabel", "dhcpOptionLabel") {
							dhcpPolMap["dhcp_option_policy_name"] = models.StripQuotes(bdCont.S("dhcpLabel", "dhcpOptionLabel", "name").String())
							dhcpPolMap["dhcp_option_policy_version"] = models.StripQuotes(bdCont.S("dhcpLabel", "dhcpOptionLabel", "version").String())
							if dhcpPolMap["dhcp_option_policy_name"] == "{}" {
								dhcpPolMap["dhcp_option_policy_name"] = nil
							}
							if dhcpPolMap["dhcp_option_policy_version"] == "{}" {
								dhcpPolMap["dhcp_option_policy_version"] = nil
							}
						}
					}
					d.Set("dhcp_policy", dhcpPolMap)
					found = true
					break
				}

			}
		}

	}

	if !found {
		d.SetId("")
	}

	log.Printf("[DEBUG] %s: Read finished successfully", d.Id())
	return nil

}

func resourceMSOTemplateBDUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Template BD: Beginning Update")
	msoClient := m.(*client.Client)

	schemaID := d.Get("schema_id").(string)
	name := d.Get("name").(string)
	displayName := d.Get("display_name").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)

	var intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, unicast_routing, arp_flooding bool
	var layer2_unknown_unicast, vrf_schema_id, vrf_template_name, virtual_mac_address, ipv6_unknown_multicast_flooding, multi_destination_flooding, unknown_multicast_flooding string

	if tempVar, ok := d.GetOk("intersite_bum_traffic"); ok {
		intersite_bum_traffic = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("optimize_wan_bandwidth"); ok {
		optimize_wan_bandwidth = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer2_stretch"); ok {
		layer2_stretch = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer3_multicast"); ok {
		layer3_multicast = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer2_unknown_unicast"); ok {
		layer2_unknown_unicast = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unknown_multicast_flooding"); ok {
		unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("multi_destination_flooding"); ok {
		multi_destination_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("ipv6_unknown_multicast_flooding"); ok {
		ipv6_unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unicast_routing"); ok {
		unicast_routing = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("virtual_mac_address"); ok {
		virtual_mac_address = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("arp_flooding"); ok {
		arp_flooding = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("vrf_schema_id"); ok {
		vrf_schema_id = tempVar.(string)
	} else {
		vrf_schema_id = schemaID
	}
	if tempVar, ok := d.GetOk("vrf_template_name"); ok {
		vrf_template_name = tempVar.(string)
	} else {
		vrf_template_name = templateName
	}
	var dhcpPolMap map[string]interface{}
	if tempVar, ok := d.GetOk("dhcp_policy"); ok {
		dhcp_policy := tempVar.(map[string]interface{})
		dhcpPolMap = make(map[string]interface{})
		dhcpPolMap["name"] = dhcp_policy["name"]
		version, err := strconv.Atoi(dhcp_policy["version"].(string))
		if err != nil {
			return err
		}
		dhcpPolMap["version"] = version

		optionName := dhcp_policy["dhcp_option_policy_name"]
		optionVersion := dhcp_policy["dhcp_option_policy_version"]
		if optionName != nil {
			if optionVersion != nil {
				dhcpOptionMap := make(map[string]interface{})
				dhcpOptionMap["name"] = optionName
				ver, err := strconv.Atoi(optionVersion.(string))
				if err != nil {
					return err
				}
				dhcpOptionMap["version"] = ver
				dhcpPolMap["dhcpOptionLabel"] = dhcpOptionMap
			} else {
				return fmt.Errorf("dhcp_option_policy_version is required with dhcp_option_policy_name")
			}
		}
	} else {
		dhcpPolMap = nil
	}

	vrfRefMap := make(map[string]interface{})
	vrfRefMap["schemaId"] = vrf_schema_id
	vrfRefMap["templateName"] = vrf_template_name
	vrfRefMap["vrfName"] = vrfName
	path := fmt.Sprintf("/templates/%s/bds/%s", templateName, name)
	bdStruct := models.NewTemplateBD("replace", path, name, displayName, layer2_unknown_unicast, unknown_multicast_flooding, multi_destination_flooding, ipv6_unknown_multicast_flooding, virtual_mac_address, intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, arp_flooding, unicast_routing, vrfRefMap, dhcpPolMap)

	_, err := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaID), bdStruct)

	if err != nil {
		return err
	}
	return resourceMSOTemplateBDRead(d, m)
}

func resourceMSOTemplateBDDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Template BD: Beginning Update")
	msoClient := m.(*client.Client)

	schemaID := d.Get("schema_id").(string)
	name := d.Get("name").(string)
	displayName := d.Get("display_name").(string)
	templateName := d.Get("template_name").(string)
	vrfName := d.Get("vrf_name").(string)

	var intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, unicast_routing, arp_flooding bool
	var layer2_unknown_unicast, vrf_schema_id, vrf_template_name, virtual_mac_address, ipv6_unknown_multicast_flooding, multi_destination_flooding, unknown_multicast_flooding string

	if tempVar, ok := d.GetOk("intersite_bum_traffic"); ok {
		intersite_bum_traffic = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("optimize_wan_bandwidth"); ok {
		optimize_wan_bandwidth = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer3_multicast"); ok {
		layer3_multicast = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("layer2_unknown_unicast"); ok {
		layer2_unknown_unicast = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unknown_multicast_flooding"); ok {
		unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("multi_destination_flooding"); ok {
		multi_destination_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("ipv6_unknown_multicast_flooding"); ok {
		ipv6_unknown_multicast_flooding = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("unicast_routing"); ok {
		unicast_routing = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("virtual_mac_address"); ok {
		virtual_mac_address = tempVar.(string)
	}
	if tempVar, ok := d.GetOk("arp_flooding"); ok {
		arp_flooding = tempVar.(bool)
	}
	if tempVar, ok := d.GetOk("vrf_schema_id"); ok {
		vrf_schema_id = tempVar.(string)
	} else {
		vrf_schema_id = schemaID
	}
	if tempVar, ok := d.GetOk("vrf_template_name"); ok {
		vrf_template_name = tempVar.(string)
	} else {
		vrf_template_name = templateName
	}
	var dhcpPolMap map[string]interface{}
	if tempVar, ok := d.GetOk("dhcp_policy"); ok {
		dhcp_policy := tempVar.(map[string]interface{})
		dhcpPolMap := make(map[string]interface{})
		dhcpPolMap["name"] = dhcp_policy["name"]
		dhcpPolMap["version"] = dhcp_policy["version"]

		optionName := dhcp_policy["dhcp_option_policy_name"]
		optionVersion := dhcp_policy["dhcp_option_policy_version"]
		if optionName != nil {
			if optionVersion != nil {
				dhcpOptionMap := make(map[string]interface{})
				dhcpOptionMap["name"] = optionName
				dhcpOptionMap["version"] = optionVersion
				dhcpPolMap["dhcpOptionLabel"] = dhcpOptionMap
			} else {
				return fmt.Errorf("dhcp_option_policy_version is required with dhcp_option_policy_name")
			}
		}
	} else {
		dhcpPolMap = nil
	}

	vrfRefMap := make(map[string]interface{})
	vrfRefMap["schemaId"] = vrf_schema_id
	vrfRefMap["templateName"] = vrf_template_name
	vrfRefMap["vrfName"] = vrfName
	path := fmt.Sprintf("/templates/%s/bds/%s", templateName, name)
	bdStruct := models.NewTemplateBD("remove", path, name, displayName, layer2_unknown_unicast, unknown_multicast_flooding, multi_destination_flooding, ipv6_unknown_multicast_flooding, virtual_mac_address, intersite_bum_traffic, optimize_wan_bandwidth, layer2_stretch, layer3_multicast, arp_flooding, unicast_routing, vrfRefMap, dhcpPolMap)

	response, err := msoClient.PatchbyID(fmt.Sprintf("api/v1/schemas/%s", schemaID), bdStruct)

	// Ignoring Error with code 141: Resource Not Found when deleting
	if err != nil && !(response.Exists("code") && response.S("code").String() == "141") {
		return err
	}
	d.SetId("")
	return nil
}
